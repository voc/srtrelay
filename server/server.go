package server

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"

import (
	"errors"
	"log"
	"strings"

	"github.com/haivision/srtgo"
	"github.com/voc/srtrelay/format"
	"github.com/voc/srtrelay/relay"
)

const (
	StreamIDSockOpt = 46

	// PacketSize = 1456
	PacketSize = 1316 // TS_UDP_LEN
)

var (
	InvalidStreamID     = errors.New("Invalid stream ID")
	InvalidMode         = errors.New("Invalid mode")
	StreamAlreadyExists = errors.New("Stream already exists")
	StreamNotExisting   = errors.New("Stream does not exist")
)

const statsPeriodMs = 2

// Server is an interface for a srt relay server
type Server interface {
	Handle(*srtgo.SrtSocket)
}

// ServerImpl implements the Server interface
type ServerImpl struct {
	ps relay.Relay
}

// NewServer creates a server
func NewServer() Server {
	ps := relay.NewRelay()
	return &ServerImpl{ps}
}

// Mode - client mode
type Mode uint8

const (
	_ Mode = iota
	ModePlay
	ModePublish
)

// ParseStreamID separates mode and stream name
func ParseStreamID(streamID string) (string, Mode, error) {
	split := strings.Split(streamID, "/")
	if len(split) != 2 {
		return "", 0, InvalidStreamID
	}
	name := split[0]
	modeStr := split[1]

	var mode Mode
	switch modeStr {
	case "play":
		mode = ModePlay
	case "publish":
		mode = ModePublish
	default:
		return "", 0, InvalidMode
	}
	return name, mode, nil
}

// Handle srt client connection
func (s *ServerImpl) Handle(sock *srtgo.SrtSocket) {
	defer sock.Close()

	streamid, err := sock.GetSockOptString(C.SRTO_STREAMID)
	if err != nil {
		log.Println(err)
		return
	}

	name, mode, err := ParseStreamID(streamid)
	if err != nil {
		log.Println(err)
		return
	}

	switch mode {
	case ModePlay:
		err = s.play(name, sock)
	case ModePublish:
		err = s.publish(name, sock)
	}
	if err != nil {
		log.Println(err)
	}
}

// play a stream from the server
func (s *ServerImpl) play(name string, sock *srtgo.SrtSocket) error {
	sub, unsubscribe, err := s.ps.Subscribe(name)
	if err != nil {
		return err
	}
	defer unsubscribe()

	log.Println("Subscribe", name)

	demux := format.NewDemuxer()
	playing := false
	for {
		buf, ok := <-sub

		// Upstream closed, drop connection
		if !ok {
			return nil
		}

		// Find synchronization point
		if !playing {
			init, err := demux.FindInit(buf)
			if err != nil {
				return err
			} else if init != nil {
				log.Println("got init", len(init))
				for i := range init {
					buf := init[i]
					sock.Write(buf, len(buf))
				}
				playing = true
			} else {
				continue
			}
		}

		// Write to socket
		sock.Write(buf, len(buf))
	}
}

// publish a stream to the server
func (s *ServerImpl) publish(name string, sock *srtgo.SrtSocket) error {
	pub, err := s.ps.Publish(name)
	if err != nil {
		return err
	}
	defer close(pub)

	log.Println("Publish", name)
	for {
		buf := make([]byte, PacketSize)
		n, err := sock.Read(buf, PacketSize)
		if err != nil {
			log.Println(err)
			return nil
		}
		// EOF
		if n == 0 {
			return nil
		}

		pub <- buf[:n]
	}
}
