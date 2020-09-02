package relay

import (
	"io"
	"log"
	"sync"

	"context"

	"github.com/asticode/go-astits"
	// "github.com/drillbits/go-ts"
	// "github.com/asticode/go-astits/astits"
)

type PubSub interface {
	Publish(string) (chan<- []byte, error)
	Subscribe(string) (<-chan []byte, error)
}

type Channel struct {
	mutex sync.Mutex
	dmx *astits.Demuxer
	subs []chan []byte
}

type PubSubImpl struct {
	mutex sync.Mutex
	channels map[string]*Channel
}

func (ch *Channel) Sub() <-chan []byte {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	sub := make(chan []byte, 10)
	ch.subs = append(ch.subs, sub)
	return sub
}

// f, _ := os.Open("/path/to/file.ts")
// defer f.Close()

// // Create the demuxer
// for {
//     // Get the next data
//     d, _ := dmx.NextData()

//     // Data is a PMT data
//     if d.PMT != nil {
//         // Loop through elementary streams
//         for _, es := range d.PMT.ElementaryStreams {
//                 fmt.Printf("Stream detected: %d\n", es.ElementaryPID)
//         }
//         return
//     }
// }

func (ch *Channel) Pub(b []byte) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	// Parse and store last PPS
	// log.Println("fwd", len(b))
	for i := range ch.subs {
		select {
			case ch.subs[i] <- b:
				continue

			// TODO: mark overflowed chan for drop
			default:
				log.Println("drop", i)
		}
	}
}

func (ch *Channel) Close() {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	for i := range ch.subs {
		close(ch.subs[i])
	}
	ch.subs = nil
}

func NewPubSub() PubSub {
	return &PubSubImpl{
		channels: make(map[string]*Channel),
	}
}

type Forwarder struct {
	ch chan []byte
}

func (f *Forwarder) Read(p []byte) (n int, err error) {
	res, ok := <- f.ch
	if !ok {
		return 0, io.EOF
	}
	copy(p, res)
	return len(res), nil
}

// Publish claims a stream name for publishing
func (s *PubSubImpl) Publish(name string) (chan<- []byte, error) {
	s.mutex.Lock()

	if _, exists := s.channels[name]; exists {
		s.mutex.Unlock()
		return nil, StreamAlreadyExists
	}

	channel := &Channel{subs: make([]chan []byte, 0, 10)}
	s.channels[name] = channel
	s.mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan []byte, 0)

	dmx := astits.New(ctx, &Forwarder{ch})

	// Setup publisher goroutine
	go func() {
		for {
			d, err := dmx.NextData()
			// buf, ok := <-ch

			// Channel closed, Teardown pubsub
			if err != nil {
				log.Println(err)
				// Need a lock on the map first to stop new subscribers
				s.mutex.Lock()
				log.Println("Removing stream", name)
				delete(s.channels, name)
				channel.Close()
				cancel()
				s.mutex.Unlock()
				return
			}

			log.Println(d)

			// Data is a PMT data
			if d.PMT != nil {
				// Loop through elementary streams
				for _, es := range d.PMT.ElementaryStreams {
					log.Printf("Stream detected: %d\n", es.ElementaryPID)
				}
				return
			}

			// Publish buf to subscribers
			// channel.Pub(buf)
		}
	}()
	return ch, nil
}

func (s *PubSubImpl) Subscribe(name string) (<-chan []byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	channel, ok := s.channels[name]
	if !ok {
		return nil, StreamNotExisting
	}
	return channel.Sub(), nil
}