package relay

import (
	"io"
	"log"
	"sync"
)

type UnsubscribeFunc func()
type PubSub interface {
	Publish(string) (chan<- []byte, error)
	Subscribe(string) (<-chan []byte, UnsubscribeFunc, error)
}

type Channel struct {
	mutex sync.Mutex
	subs []chan []byte
}

type PubSubImpl struct {
	mutex sync.Mutex
	channels map[string]*Channel
}


func (ch *Channel) Sub() (<-chan []byte, UnsubscribeFunc) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	sub := make(chan []byte, 80) // about 300ms at 3Mbit/s
	ch.subs = append(ch.subs, sub)

	var unsub UnsubscribeFunc = func() {
		ch.mutex.Lock()
		defer ch.mutex.Unlock()
		var idx int
		for i := range ch.subs {
			if ch.subs[i] == sub {
				idx = i
				break
			}
		}
		ch.subs = append(ch.subs[:idx], ch.subs[idx+1:]...)
		log.Println("unsub", idx)
	}
	return sub, unsub
}

func (ch *Channel) Pub(b []byte) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	for i := range ch.subs {
		select {
			case ch.subs[i] <- b:
				continue

			// TODO: mark overflowed chan for drop
			default:
				close(ch.subs[i])
				log.Println("dropping client", i)
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

	ch := make(chan []byte, 0)

	// Setup publisher goroutine
	go func() {
		for {
			buf, ok := <-ch

			// Channel closed, Teardown pubsub
			if !ok {
				// Need a lock on the map first to stop new subscribers
				s.mutex.Lock()
				log.Println("Removing stream", name)
				delete(s.channels, name)
				channel.Close()
				s.mutex.Unlock()
				return
			}

			// Publish buf to subscribers
			channel.Pub(buf)
		}
	}()
	return ch, nil
}

func (s *PubSubImpl) Subscribe(name string) (<-chan []byte, UnsubscribeFunc, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	channel, ok := s.channels[name]
	if !ok {
		return nil, nil, StreamNotExisting
	}
	ch, unsub := channel.Sub()
	return ch, unsub, nil
}