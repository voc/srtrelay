package relay

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/voc/srtrelay/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const relaySubsystem = "relay"

var (
	activeClients = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(config.MetricsNamespace, relaySubsystem, "active_clients"),
			Help: "The number of active clients per channel",
		},
		[]string{"channel_name"},
	)
	channelCreatedTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(config.MetricsNamespace, relaySubsystem, "created_timestamp_seconds"),
			Help: "The UNIX timestamp when the channel was created",
		},
		[]string{"channel_name"},
	)
)

type UnsubscribeFunc func()

type Channel struct {
	name       string
	mutex      sync.Mutex
	subs       Subs
	maxPackets uint

	// statistics
	clients atomic.Value
	created time.Time

	// Prometheus metrics.
	activeClients    prometheus.Gauge
	createdTimestamp prometheus.Gauge
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

func NewChannel(name string, maxPackets uint) *Channel {
	channelActiveClients := activeClients.WithLabelValues(name)
	ch := &Channel{
		name:          name,
		subs:          make([]chan []byte, 0, 10),
		maxPackets:    maxPackets,
		created:       time.Now(),
		activeClients: channelActiveClients,
	}
	ch.clients.Store(0)
	ch.createdTimestamp = channelCreatedTimestamp.WithLabelValues(name)
	ch.createdTimestamp.Set(float64(ch.created.UnixNano()) / 1000000.0)
	return ch
}

// Sub subscribes to a channel
func (ch *Channel) Sub() (<-chan []byte, UnsubscribeFunc) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	sub := make(chan []byte, ch.maxPackets)
	ch.subs = append(ch.subs, sub)
	ch.clients.Store(len(ch.subs))
	ch.activeClients.Inc()

	var unsub UnsubscribeFunc = func() {
		ch.mutex.Lock()
		defer ch.mutex.Unlock()

		// Channel already closed, just skip unsub
		if ch.subs == nil {
			return
		}

		ch.subs = ch.subs.Remove(sub)
		ch.clients.Store(len(ch.subs))
		ch.activeClients.Dec()
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
			log.Println("dropping overflowing client", i)
		}
	}
	for _, sub := range toRemove {
		ch.subs = ch.subs.Remove(sub)
		ch.activeClients.Dec()
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
	activeClients.DeleteLabelValues(ch.name)
	channelCreatedTimestamp.DeleteLabelValues(ch.name)
}

func (ch *Channel) Stats() Stats {
	return Stats{
		clients: ch.clients.Load().(int),
		created: ch.created,
	}
}
