package relay

import (
	"log"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/voc/srtrelay/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const relaySubsystem = "relay"

var (
	activeClients = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(metrics.Namespace, relaySubsystem, "active_clients"),
			Help: "The number of active clients per channel",
		},
		[]string{"channel_name"},
	)
	channelCreatedTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(metrics.Namespace, relaySubsystem, "created_timestamp_seconds"),
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
	closed     chan struct{}

	// statistics
	clients atomic.Value
	created time.Time

	// Prometheus metrics.
	activeClients    prometheus.Gauge
	createdTimestamp prometheus.Gauge
}
type Subs []*Subscriber

type Stats struct {
	clients int
	created time.Time
}

// Remove single subscriber
func (subs Subs) Remove(sub chan []byte) Subs {
	idx := -1
	for i := range subs {
		if subs[i].ch == sub {
			idx = i
			break
		}
	}

	// subscriber was already removed
	if idx < 0 {
		return subs
	}

	close(sub)
	return slices.Delete(subs, idx, idx+1)
}

func NewChannel(name string, maxPackets uint) *Channel {
	channelActiveClients := activeClients.WithLabelValues(name)
	ch := &Channel{
		name:          name,
		subs:          make(Subs, 0, 10),
		maxPackets:    maxPackets,
		created:       time.Now(),
		activeClients: channelActiveClients,
		closed:        make(chan struct{}),
	}
	ch.clients.Store(0)
	ch.createdTimestamp = channelCreatedTimestamp.WithLabelValues(name)
	ch.createdTimestamp.Set(float64(ch.created.UnixNano()) / 1000000.0)
	return ch
}

// Sub subscribes to a channel
func (ch *Channel) Sub() (*Subscriber, UnsubscribeFunc) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	sub := &Subscriber{
		ch:         make(chan []byte, ch.maxPackets),
		chanClosed: ch.closed,
	}
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

		ch.subs = ch.subs.Remove(sub.ch)
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
		case ch.subs[i].ch <- b:
			continue

		// Remember overflowed chans for drop
		default:
			toRemove = append(toRemove, ch.subs[i])
			log.Println("dropping overflowing client", i)
		}
	}
	for _, sub := range toRemove {
		ch.subs = ch.subs.Remove(sub.ch)
		ch.activeClients.Dec()
	}
	ch.clients.Store(len(ch.subs))
}

// Close closes a channel
func (ch *Channel) Close() {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	close(ch.closed)
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
