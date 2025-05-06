package srt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/netip"
	"sync"
	"time"

	gosrt "github.com/datarhei/gosrt"

	"github.com/voc/srtrelay/auth"
	"github.com/voc/srtrelay/format"
	"github.com/voc/srtrelay/relay"
	"github.com/voc/srtrelay/stream"
)

type Config struct {
	Server ServerConfig
	Relay  relay.RelayConfig
}

type ServerConfig struct {
	Addresses     []netip.AddrPort
	PublicAddress string
	LatencyMs     uint
	LossMaxTTL    uint32
	PacketSize    uint32
	Auth          auth.Authenticator
	SyncClients   bool
}

type Server struct {
	config    *ServerConfig
	listeners []gosrt.Listener
	relay     *relay.Relay

	mutex sync.Mutex
	conns map[*srtConn]bool
	done  sync.WaitGroup
}

// NewServer creates a server
func NewServer(config *Config) *Server {
	r := relay.NewRelay(&config.Relay)
	return &Server{
		relay:  r,
		config: &config.Server,
		conns:  make(map[*srtConn]bool),
	}
}

// Listen sets up a SRT socket in listen mode
func (s *Server) Listen(ctx context.Context) error {
	for _, address := range s.config.Addresses {
		listener, err := s.listenAt(ctx, address)
		if err != nil {
			return err
		}
		log.Printf("SRT Listening on %s\n", address)
		s.listeners = append(s.listeners, listener)
	}

	return nil
}

func (s *Server) Addresses() []net.Addr {
	addrs := make([]net.Addr, 0, len(s.listeners))
	for _, listener := range s.listeners {
		addrs = append(addrs, listener.Addr())
	}
	return addrs
}

// Wait blocks until listening sockets have been closed
func (s *Server) Wait() {
	s.done.Wait()
}

func (s *Server) listenAt(ctx context.Context, addr netip.AddrPort) (gosrt.Listener, error) {
	conf := gosrt.DefaultConfig()
	conf.Latency = time.Duration(s.config.LatencyMs) * time.Millisecond
	conf.PayloadSize = s.config.PacketSize
	conf.LossMaxTTL = s.config.LossMaxTTL
	ln, err := gosrt.Listen("srt", addr.String(), conf)
	if err != nil {
		return nil, err
	}

	s.done.Add(2)
	// server socket closer
	go func() {
		defer s.done.Done()
		<-ctx.Done()
		ln.Close()
	}()

	// accept loop
	go func() {
		defer s.done.Done()
		for {
			req, err := ln.Accept2()
			if err != nil {
				// exit silently on close
				if errors.Is(err, gosrt.ErrListenerClosed) {
					return
				}
				log.Println("accept failed", err)
			}

			if reason, ok := s.shouldAccept(req); !ok {
				req.Reject(reason)
				continue
			}

			conn, err := req.Accept()
			if err != nil {
				log.Println("accept failed", err)
				continue
			}
			go s.Handle(ctx, conn)
		}
	}()
	return ln, nil
}

func (s *Server) shouldAccept(req gosrt.ConnRequest) (gosrt.RejectionReason, bool) {
	var streamid stream.StreamID

	// Parse stream id
	if err := streamid.FromString(req.StreamId()); err != nil {
		log.Println(err)
		return gosrt.REJ_PEER, false
	}

	// Check authentication
	if !s.config.Auth.Authenticate(streamid) {
		log.Printf("%s - Stream '%s' access denied\n", req.RemoteAddr(), streamid)
		return gosrt.REJX_UNAUTHORIZED, false
	}

	return 0, true
}

// SRTConn wraps an srtsocket with additional state
type srtConn struct {
	log      *slog.Logger
	socket   relaySocket
	streamid *stream.StreamID
}

type relaySocket interface {
	io.Reader
	io.Writer
	RemoteAddr() net.Addr
	Close() error
	Stats(*gosrt.Statistics)
}

// Handle srt client connection
func (s *Server) Handle(ctx context.Context, conn gosrt.Conn) {
	var streamid stream.StreamID
	defer conn.Close()

	// Parse stream id
	if err := streamid.FromString(conn.StreamId()); err != nil {
		log.Println(err)
		return
	}

	myconn := &srtConn{
		log:      slog.With("addr", conn.RemoteAddr(), "stream", streamid.Name(), "mode", streamid.Mode()),
		socket:   conn,
		streamid: &streamid,
	}

	subctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s.registerForStats(subctx, myconn)

	var err error
	switch streamid.Mode() {
	case stream.ModePlay:
		err = s.play(myconn)
	case stream.ModePublish:
		err = s.publish(myconn)
	}
	if err != nil {
		myconn.log.Info("closing", "error", err)
	}
}

// play a stream from the server
func (s *Server) play(conn *srtConn) error {
	sub, unsubscribe, err := s.relay.Subscribe(conn.streamid.Name())
	if err != nil {
		return err
	}
	defer unsubscribe()
	conn.log.Info("play")

	demux := format.NewDemuxer()
	playing := !s.config.SyncClients
	for {
		buf, err := sub.Read()
		// Upstream closed, drop connection
		if err != nil {
			conn.log.Info("disconnecting", "error", err)
			return nil
		}

		// Find initial synchronization point
		// TODO: implement timeout for sync
		if !playing {
			init, err := demux.FindInit(buf)
			if err != nil {
				return err
			} else if init != nil {
				for i := range init {
					buf := init[i]
					_, err := conn.socket.Write(buf)
					if err != nil {
						return fmt.Errorf("write init: %w", err)
					}
				}
				playing = true
			}
			continue
		}

		// Write to socket
		_, err = conn.socket.Write(buf)
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
}

// publish a stream to the server
func (s *Server) publish(conn *srtConn) error {
	pub, err := s.relay.Publish(conn.streamid.Name())
	if err != nil {
		return err
	}
	defer close(pub)
	conn.log.Info("publish")

	buf := make([]byte, s.config.PacketSize)
	for {
		n, err := conn.socket.Read(buf)

		fwd := make([]byte, n)
		copy(fwd, buf[:n])

		// Push read buffers to all clients via the publish channel
		if n > 0 {
			pub <- fwd
		}

		if err != nil {
			return err
		}
	}
}

func (s *Server) registerForStats(ctx context.Context, conn *srtConn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.conns[conn] = true

	go func() {
		<-ctx.Done()

		s.mutex.Lock()
		defer s.mutex.Unlock()

		delete(s.conns, conn)
	}()
}

func (s *Server) GetStatistics() []*relay.StreamStatistics {
	streams := s.relay.GetStatistics()
	for _, st := range streams {
		st.URL = fmt.Sprintf("srt://%s?streamid=#!::m=request,r=%s", s.config.PublicAddress, st.Name) // New format
	}
	return streams
}

type SocketStatistics struct {
	Address  string                      `json:"address"`
	StreamID string                      `json:"stream_id"`
	Stats    gosrt.StatisticsAccumulated `json:"stats"`
}

func (s *Server) GetSocketStatistics() []*SocketStatistics {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var stats gosrt.Statistics
	statistics := make([]*SocketStatistics, 0, len(s.conns))
	for conn := range s.conns {
		conn.socket.Stats(&stats)
		statistics = append(statistics, &SocketStatistics{
			Address:  conn.socket.RemoteAddr().String(),
			StreamID: conn.streamid.String(),
			Stats:    stats.Accumulated,
		})
	}

	return statistics
}
