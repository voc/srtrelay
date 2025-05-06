package relay

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrStreamAlreadyExists = errors.New("stream already exists")
	ErrStreamNotExisting   = errors.New("stream does not exist")
)

type RelayConfig struct {
	BufferSize uint
	PacketSize uint
}

type StreamStatistics struct {
	Name    string    `json:"name"`
	URL     string    `json:"url"`
	Clients int       `json:"clients"`
	Created time.Time `json:"created"`
}

// Relay represents a multi-channel stream relay
type Relay struct {
	mutex    sync.Mutex
	channels map[string]*Channel
	config   *RelayConfig
}

// NewRelay creates a relay
func NewRelay(config *RelayConfig) *Relay {
	if config.BufferSize == 0 {
		config.BufferSize = 384000 // 1s @ 3Mbits/s
	}
	if config.PacketSize == 0 {
		config.PacketSize = 1316 // default packet size
	}
	return &Relay{
		channels: make(map[string]*Channel),
		config:   config,
	}
}

// Publish claims a stream name for publishing
func (s *Relay) Publish(name string) (chan<- []byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.channels[name]; exists {
		return nil, ErrStreamAlreadyExists
	}

	channel := NewChannel(name, s.config.BufferSize/s.config.PacketSize)
	s.channels[name] = channel

	ch := make(chan []byte)

	// Setup publisher goroutine
	go func() {
		for {
			buf, ok := <-ch

			// Channel closed, Teardown pubsub
			if !ok {
				// Need a lock on the map first to stop new subscribers
				s.mutex.Lock()
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

type Subscriber struct {
	ch         chan []byte
	chanClosed <-chan struct{}
}

var (
	errUpstreamClosed = errors.New("upstream closed")
	errLateSubscriber = errors.New("late subscriber")
)

func (s *Subscriber) Read() ([]byte, error) {
	select {
	case <-s.chanClosed:
		return nil, errUpstreamClosed
	case buf, ok := <-s.ch:
		if !ok {
			return nil, errLateSubscriber
		}
		return buf, nil
	}
}

// Subscribe subscribes to a stream by name
func (s *Relay) Subscribe(name string) (*Subscriber, UnsubscribeFunc, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	channel, ok := s.channels[name]
	if !ok {
		return nil, nil, ErrStreamNotExisting
	}
	ch, unsub := channel.Sub()
	return ch, unsub, nil
}

func (s *Relay) GetStatistics() []*StreamStatistics {
	statistics := make([]*StreamStatistics, 0)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for name, channel := range s.channels {
		stats := channel.Stats()
		statistics = append(statistics, &StreamStatistics{
			Name:    name,
			Clients: stats.clients,
			Created: stats.created,
		})
	}
	return statistics
}
