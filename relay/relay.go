package relay

import (
	"errors"
	"log"
	"sync"
)

var (
	StreamAlreadyExists = errors.New("Stream already exists")
	StreamNotExisting   = errors.New("Stream does not exist")
)

type StreamStat struct {
	name    string
	clients int
}

type RelayConfig struct {
	Buffersize uint
}

type Relay interface {
	Publish(string) (chan<- []byte, error)
	Subscribe(string) (<-chan []byte, UnsubscribeFunc, error)
	GetStatistics(string) []*StreamStat
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
		return nil, StreamAlreadyExists
	}

	channel := NewChannel(s.config.Buffersize)
	s.channels[name] = channel

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

// Subscribe subscribes to a stream by name
func (s *RelayImpl) Subscribe(name string) (<-chan []byte, UnsubscribeFunc, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	channel, ok := s.channels[name]
	if !ok {
		return nil, nil, StreamNotExisting
	}
	ch, unsub := channel.Sub()
	return ch, unsub, nil
}

func (s *RelayImpl) GetStatistics(string) []*StreamStat {
	return nil
}
