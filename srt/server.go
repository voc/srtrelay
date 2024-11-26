package srt

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/haivision/srtgo"
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
	Addresses     []string
	PublicAddress string
	Latency       uint
	LossMaxTTL    uint
	Auth          auth.Authenticator
	SyncClients   bool
	ListenBacklog int
}

// Server is an interface for a srt relay server
type Server interface {
	Listen(context.Context) error
	Wait()
	Handle(context.Context, *srtgo.SrtSocket, *net.UDPAddr)
	GetStatistics() []*relay.StreamStatistics
	GetSocketStatistics() []*SocketStatistics
}

// ServerImpl implements the Server interface
type ServerImpl struct {
	config *ServerConfig
	relay  relay.Relay

	mutex sync.Mutex
	conns map[*srtConn]bool
	done  sync.WaitGroup

	pool *sync.Pool
}

// NewServer creates a server
func NewServer(config *Config) *ServerImpl {
	r := relay.NewRelay(&config.Relay)
	return &ServerImpl{
		relay:  r,
		config: &config.Server,
		conns:  make(map[*srtConn]bool),
		pool:   newBufferPool(config.Relay.PacketSize),
	}
}

// Listen sets up a SRT socket in listen mode
func (s *ServerImpl) Listen(ctx context.Context) error {
	for _, address := range s.config.Addresses {
		host, portString, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}

		port, err := strconv.ParseUint(portString, 10, 16)
		if err != nil {
			return err
		}

		var addresses []string
		if len(host) > 0 {
			addresses, err = net.LookupHost(host)
			if err != nil {
				return err
			}
		} else {
			addresses = []string{"::"}
		}

		for _, address := range addresses {
			err := s.listenAt(ctx, address, uint16(port))
			if err != nil {
				return err
			}
			log.Printf("SRT Listening on %s:%d\n", address, port)
		}
	}

	return nil
}

// Wait blocks until listening sockets have been closed
func (s *ServerImpl) Wait() {
	s.done.Wait()
}

func (s *ServerImpl) listenCallback(socket *srtgo.SrtSocket, version int, addr *net.UDPAddr, idstring string) bool {
	var streamid stream.StreamID

	// Parse stream id
	if err := streamid.FromString(idstring); err != nil {
		log.Println(err)
		return false
	}

	// Check authentication
	if !s.config.Auth.Authenticate(streamid) {
		log.Printf("%s - Stream '%s' access denied\n", addr, streamid)
		if err := socket.SetRejectReason(srtgo.RejectionReasonUnauthorized); err != nil {
			log.Printf("Error rejecting stream: %s", err)
		}
		return false
	}

	return true
}

func (s *ServerImpl) listenAt(ctx context.Context, host string, port uint16) error {
	options := make(map[string]string)
	options["blocking"] = "0"
	options["transtype"] = "live"
	options["latency"] = strconv.Itoa(int(s.config.Latency))

	sck := srtgo.NewSrtSocket(host, port, options)
	if err := sck.SetSockOptInt(srtgo.SRTO_LOSSMAXTTL, int(s.config.LossMaxTTL)); err != nil {
		log.Printf("Error settings lossmaxttl: %s", err)
	}
	sck.SetListenCallback(s.listenCallback)
	err := sck.Listen(s.config.ListenBacklog)
	if err != nil {
		return fmt.Errorf("Listen failed for %v:%v : %v", host, port, err)
	}

	s.done.Add(2)
	// server socket closer
	go func() {
		defer s.done.Done()
		<-ctx.Done()
		sck.Close()
	}()

	// accept loop
	go func() {
		defer s.done.Done()
		for {
			sck.SetReadDeadline(time.Now().Add(time.Millisecond * 300))
			sock, addr, err := sck.Accept()
			if err != nil {
				if errors.Is(err, &srtgo.SrtEpollTimeout{}) {
					continue
				}
				// exit silently if context closed
				select {
				case <-ctx.Done():
					return
				default:
				}
				log.Println("accept failed", err)
				continue
			}
			go s.Handle(ctx, sock, addr)
		}
	}()
	return nil
}

// SRTConn wraps an srtsocket with additional state
type srtConn struct {
	socket   relaySocket
	address  string
	streamid *stream.StreamID
}

type relaySocket interface {
	io.Reader
	io.Writer
	Close()
	Stats() (*srtgo.SrtStats, error)
}

// Handle srt client connection
func (s *ServerImpl) Handle(ctx context.Context, sock *srtgo.SrtSocket, addr *net.UDPAddr) {
	var streamid stream.StreamID
	defer sock.Close()

	idstring, err := sock.GetSockOptString(C.SRTO_STREAMID)
	if err != nil {
		log.Println(err)
		return
	}

	// Parse stream id
	if err := streamid.FromString(idstring); err != nil {
		log.Println(err)
		return
	}

	conn := &srtConn{
		socket:   sock,
		address:  addr.String(),
		streamid: &streamid,
	}

	subctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s.registerForStats(subctx, conn)

	switch streamid.Mode() {
	case stream.ModePlay:
		err = s.play(conn)
	case stream.ModePublish:
		err = s.publish(conn)
	}
	if err != nil {
		log.Printf("%s - %s - %v", conn.address, conn.streamid.Name(), err)
	}
}

// play a stream from the server
func (s *ServerImpl) play(conn *srtConn) error {
	sub, unsubscribe, err := s.relay.Subscribe(conn.streamid.Name())
	if err != nil {
		return err
	}
	defer unsubscribe()
	log.Printf("%s - play %s\n", conn.address, conn.streamid.Name())

	demux := format.NewDemuxer()
	playing := !s.config.SyncClients
	for {
		buf, ok := <-sub

		buffered := len(sub)
		if buffered > cap(sub)/2 {
			log.Printf("%s - %s - %d packets late in buffer\n", conn.address, conn.streamid.Name(), len(sub))
		}

		// Upstream closed, drop connection
		if !ok {
			log.Printf("%s - %s dropped", conn.address, conn.streamid.Name())
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
						return err
					}
				}
				playing = true
			}
			continue
		}

		// Write to socket
		_, err := conn.socket.Write(buf)
		if err != nil {
			return err
		}
	}
}

// publish a stream to the server
func (s *ServerImpl) publish(conn *srtConn) error {
	pub, err := s.relay.Publish(conn.streamid.Name())
	if err != nil {
		return err
	}
	defer close(pub)
	log.Printf("%s - publish %s\n", conn.address, conn.streamid.Name())

	for {
		// Get buffer from pool and return sometime after use
		buf := s.pool.Get().(*[]byte)
		runtime.SetFinalizer(buf, func(buf *[]byte) {
			s.pool.Put(buf)
		})

		n, err := conn.socket.Read(*buf)

		// Push read buffers to all clients via the publish channel
		if n > 0 {
			pub <- (*buf)[:n]
		}

		if err != nil {
			return err
		}
	}
}

func (s *ServerImpl) registerForStats(ctx context.Context, conn *srtConn) {
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

func (s *ServerImpl) GetStatistics() []*relay.StreamStatistics {
	streams := s.relay.GetStatistics()
	for _, st := range streams {
		//stream.URL = fmt.Sprintf("srt://%s?streamid=#!::m=request,r=%s", s.config.PublicAddress, stream.Name)
		st.URL = fmt.Sprintf("srt://%s?streamid=play/%s", s.config.PublicAddress, st.Name)
	}
	return streams
}

type SocketStatistics struct {
	Address  string          `json:"address"`
	StreamID string          `json:"stream_id"`
	Stats    *srtgo.SrtStats `json:"stats"`
}

func (s *ServerImpl) GetSocketStatistics() []*SocketStatistics {
	statistics := make([]*SocketStatistics, 0)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for conn := range s.conns {
		srtStats, err := conn.socket.Stats()
		if err != nil {
			log.Printf("%s - error getting stats %s\n", conn.address, err)
			continue
		}
		statistics = append(statistics, &SocketStatistics{
			Address:  conn.address,
			StreamID: conn.streamid.String(),
			Stats:    srtStats,
		})
	}

	return statistics
}
