package server

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/haivision/srtgo"
	"github.com/voc/srtrelay/auth"
	"github.com/voc/srtrelay/format"
	"github.com/voc/srtrelay/relay"
	"github.com/voc/srtrelay/stream"
)

const (
	// Make this configurable? max is 1456
	PacketSize = 1316 // TS_UDP_LEN
)

type Config struct {
	Server ServerConfig
	Relay  relay.RelayConfig
}

type ServerConfig struct {
	Address string
	Port    uint16
	Latency uint
	Auth    auth.Authenticator
}

// Server is an interface for a srt relay server
type Server interface {
	Listen(context.Context) error
	Handle(*srtgo.SrtSocket, *net.UDPAddr)
}

// ServerImpl implements the Server interface
type ServerImpl struct {
	config *ServerConfig
	relay  relay.Relay
}

// NewServer creates a server
func NewServer(config *Config) Server {
	r := relay.NewRelay(&config.Relay)
	return &ServerImpl{
		relay:  r,
		config: &config.Server,
	}
}

// Listen sets up a SRT socket in listen mode
func (s *ServerImpl) Listen(ctx context.Context) error {
	options := make(map[string]string)
	options["blocking"] = "0"
	options["transtype"] = "live"
	options["latency"] = strconv.Itoa(int(s.config.Latency))

	sck := srtgo.NewSrtSocket(s.config.Address, s.config.Port, options)
	err := sck.Listen(1)
	if err != nil {
		return fmt.Errorf("Listen failed: %v", err)
	}

	go func() {
		<-ctx.Done()
		sck.Close()
	}()

	go func() {
		for {
			sock, addr, err := sck.Accept()
			if err != nil {
				// exit silently if context closed
				select {
				case <-ctx.Done():
					return
				default:
				}
				log.Fatalln("accept failed", err)
			}
			go s.Handle(sock, addr)
		}
	}()
	return nil
}

// SRTConn wraps an srtsocket with additional state
type srtConn struct {
	socket   *srtgo.SrtSocket
	address  *net.UDPAddr
	streamid *stream.StreamID
}

// Handle srt client connection
func (s *ServerImpl) Handle(sock *srtgo.SrtSocket, addr *net.UDPAddr) {
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

	// Check authentication
	if !s.config.Auth.Authenticate(streamid) {
		log.Printf("%s - Stream '%s' access denied\n", addr, streamid)
		return
	}

	conn := &srtConn{
		socket:   sock,
		address:  addr,
		streamid: &streamid,
	}

	switch streamid.Mode() {
	case stream.ModePlay:
		err = s.play(conn)
	case stream.ModePublish:
		err = s.publish(conn)
	}
	if err != nil {
		log.Printf("%s - %v", conn.address, err)
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
	playing := false
	for {
		buf, ok := <-sub

		buffered := len(sub)
		if buffered > 144 {
			log.Printf("%s - %d packets late in buffer\n", conn.address, len(sub))
		}

		// Upstream closed, drop connection
		if !ok {
			log.Println("dropping", conn.address)
			return nil
		}

		// Find synchronization point
		// TODO: implement timeout for initial sync
		if !playing {
			init, err := demux.FindInit(buf)
			if err != nil {
				return err
			} else if init != nil {
				for i := range init {
					buf := init[i]
					conn.socket.Write(buf, len(buf))
				}
				playing = true
			} else {
				continue
			}
		}

		// Write to socket
		_, err := conn.socket.Write(buf, len(buf))
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
		// Push read buffers to all clients via the publish channel
		// a ringbuffer would probably be more efficient
		buf := make([]byte, PacketSize)
		n, err := conn.socket.Read(buf, PacketSize)
		if err != nil {
			return err
		}

		// handle EOF
		if n == 0 {
			return nil
		}

		pub <- buf[:n]
	}
}
