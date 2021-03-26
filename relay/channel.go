package relay

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type UnsubscribeFunc func()

type Channel struct {
	mutex      sync.Mutex
	subs       Subs
	buffersize uint

	// statistics
	clients atomic.Value
	created time.Time
}
type Subs []chan []byte

type Stats struct {
	clients int
	created time.Time
}

// Remove single subscriber
func (subs Subs) Remove(sub chan []byte) Subs {
	idx := -1
	for i := range subs {
		if subs[i] == sub {
			idx = i
			break
		}
	}

	// subscriber was already removed
	if idx < 0 {
		return subs
	}

	defer close(sub)

	subs[idx] = subs[len(subs)-1] // Copy last element to index i.
	subs[len(subs)-1] = nil       // Erase last element (write zero value).
	return subs[:len(subs)-1]     // Truncate slice.
}

func NewChannel(buffersize uint) *Channel {
	ch := &Channel{
		subs:       make([]chan []byte, 0, 10),
		buffersize: buffersize,
		created:    time.Now(),
	}
	ch.clients.Store(0)
	return ch
}

// Sub subscribes to a channel
func (ch *Channel) Sub() (<-chan []byte, UnsubscribeFunc) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	channelbuffer := ch.buffersize / 1316
	sub := make(chan []byte, channelbuffer)
	ch.subs = append(ch.subs, sub)
	ch.clients.Store(len(ch.subs))

	var unsub UnsubscribeFunc = func() {
		ch.mutex.Lock()
		defer ch.mutex.Unlock()

		// Channel already closed, just skip unsub
		if ch.subs == nil {
			return
		}

		ch.subs = ch.subs.Remove(sub)
		ch.clients.Store(len(ch.subs))
	}
	return sub, unsub
}

// Pub publishes a packet to a channel
func (ch *Channel) Pub(b []byte) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	toRemove := make(Subs, 0, 5)
	for i := range ch.subs {
		select {
		case ch.subs[i] <- b:
			continue

		// Remember overflowed chans for drop
		default:
			toRemove = append(toRemove, ch.subs[i])
			log.Println("dropping client", i)
		}
	}
	for _, sub := range toRemove {
		ch.subs = ch.subs.Remove(sub)
	}
	ch.clients.Store(len(ch.subs))
}

// Close closes a channel
func (ch *Channel) Close() {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	for i := range ch.subs {
		close(ch.subs[i])
	}
	ch.subs = nil
}

func (ch *Channel) Stats() Stats {
	return Stats{
		clients: ch.clients.Load().(int),
		created: ch.created,
	}
}
