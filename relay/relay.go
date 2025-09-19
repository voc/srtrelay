package relay

import (
	"errors"
	"log"
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

type Relay interface {
	Publish(string) (chan<- []byte, error)
	Subscribe(string) (<-chan []byte, UnsubscribeFunc, error)
	GetStatistics() []*StreamStatistics
	ChannelExists(name string) bool
}

type StreamStatistics struct {
	Name    string    `json:"name"`
	URL     string    `json:"url"`
	Clients int       `json:"clients"`
	Created time.Time `json:"created"`
}

// RelayImpl represents a multi-channel stream relay
type RelayImpl struct {
	mutex    sync.Mutex
	channels map[string]*Channel
	config   *RelayConfig
}

// NewRelay creates a relay
func NewRelay(config *RelayConfig) Relay {
	return &RelayImpl{
		channels: make(map[string]*Channel),
		config:   config,
	}
}

// Publish claims a stream name for publishing
func (s *RelayImpl) Publish(name string) (chan<- []byte, error) {
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
				log.Println("Unpublished stream", name)
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

// Subscribe subscribes to a stream by name
func (s *RelayImpl) Subscribe(name string) (<-chan []byte, UnsubscribeFunc, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	channel, ok := s.channels[name]
	if !ok {
		return nil, nil, ErrStreamNotExisting
	}
	ch, unsub := channel.Sub()
	return ch, unsub, nil
}

func (s *RelayImpl) GetStatistics() []*StreamStatistics {
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

func (s *RelayImpl) ChannelExists(name string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exists := s.channels[name]
	return exists
}
