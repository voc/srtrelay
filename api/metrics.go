package api

import (
	"github.com/voc/srtrelay/srt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace    = "srtrelay"
	srtSubsystem = "srt"
)

var (
	activeSocketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, srtSubsystem, "active_sockets"),
		"The number of active SRT sockets",
		nil, nil,
	)

	// Metrics from: https://pkg.go.dev/github.com/haivision/srtgo#SrtStats
	pktSentTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, srtSubsystem, "packets_sent_total"),
		"total number of sent data packets, including retransmissions",
		[]string{"address", "stream_id"}, nil,
	)

	pktRecvTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, srtSubsystem, "packets_received_total"),
		"total number of received packets",
		[]string{"address", "stream_id"}, nil,
	)

	// TODO: Add metrics for additional SrtStats.
	//
	// PktSndLossTotal    int   // total number of lost packets (sender side)
	// PktRcvLossTotal    int   // total number of lost packets (receiver side)
	// PktRetransTotal    int   // total number of retransmitted packets
	// PktSentACKTotal    int   // total number of sent ACK packets
	// PktRecvACKTotal    int   // total number of received ACK packets
	// PktSentNAKTotal    int   // total number of sent NAK packets
	// PktRecvNAKTotal    int   // total number of received NAK packets
	// UsSndDurationTotal int64 // total time duration when UDT is sending data (idle time exclusive)

	// PktSndDropTotal      int   // number of too-late-to-send dropped packets
	// PktRcvDropTotal      int   // number of too-late-to play missing packets
	// PktRcvUndecryptTotal int   // number of undecrypted packets
	// ByteSentTotal        int64 // total number of sent data bytes, including retransmissions
	// ByteRecvTotal        int64 // total number of received bytes
	// ByteRcvLossTotal     int64 // total number of lost bytes

	// ByteRetransTotal      int64 // total number of retransmitted bytes
	// ByteSndDropTotal      int64 // number of too-late-to-send dropped bytes
	// ByteRcvDropTotal      int64 // number of too-late-to play missing bytes (estimate based on average packet size)
	// ByteRcvUndecryptTotal int64 // number of undecrypted bytes
)

// Exporter collects metrics. It implements prometheus.Collector.
type Exporter struct {
	server srt.Server
}

func NewExporter(s srt.Server) *Exporter {
	e := Exporter{server: s}
	return &e
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- activeSocketsDesc
	ch <- pktSentTotalDesc
	ch <- pktRecvTotalDesc
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	stats := e.server.GetSocketStatistics()
	ch <- prometheus.MustNewConstMetric(activeSocketsDesc, prometheus.GaugeValue, float64(len(stats)))
	for _, stat := range stats {
		ch <- prometheus.MustNewConstMetric(pktSentTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktSentTotal), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktRecvTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktRecvTotal), stat.Address, stat.StreamID)
	}
}
