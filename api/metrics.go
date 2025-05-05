package api

import (
	"github.com/voc/srtrelay/internal/metrics"
	"github.com/voc/srtrelay/srt"

	"github.com/prometheus/client_golang/prometheus"
)

const srtSubsystem = "srt"

var (
	activeSocketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "active_sockets"),
		"The number of active SRT sockets",
		nil, nil,
	)

	pktSentTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "sent_packets_total"),
		"total number of sent data packets, including retransmissions",
		[]string{"address", "stream_id"}, nil,
	)

	pktRecvTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_packets_total"),
		"total number of received packets",
		[]string{"address", "stream_id"}, nil,
	)

	pktSndLossTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "sent_lost_packets_total"),
		"total number of lost packets (sender side)",
		[]string{"address", "stream_id"}, nil,
	)

	pktRcvLossTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_lost_packets_total"),
		"total number of lost packets (receive_side)",
		[]string{"address", "stream_id"}, nil,
	)

	pktRetransTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "retransmitted_packets_total"),
		"total number of retransmitted packets",
		[]string{"address", "stream_id"}, nil,
	)

	pktSentACKTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "sent_ack_packets_total"),
		"total number of sent ACK packets",
		[]string{"address", "stream_id"}, nil,
	)

	pktRecvACKTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_ack_packets_total"),
		"total number of received ACK packets",
		[]string{"address", "stream_id"}, nil,
	)

	pktSentNAKTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "sent_nak_packets_total"),
		"total number of received NAK packets",
		[]string{"address", "stream_id"}, nil,
	)

	pktRecvNAKTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_nak_packets_total"),
		"total number of received NAK packets",
		[]string{"address", "stream_id"}, nil,
	)

	sndDurationTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "udt_sent_duration_seconds_total"),
		"total time duration when UDT is sending data (idle time exclusive)",
		[]string{"address", "stream_id"}, nil,
	)

	pktSndDropTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "sent_dropped_packets_total"),
		"number of too-late-to-send dropped packets",
		[]string{"address", "stream_id"}, nil,
	)

	pktRcvDropTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_dropped_packets_total"),
		"number of too-late-to play missing packets",
		[]string{"address", "stream_id"}, nil,
	)

	pktRcvUndecryptTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_undecrypted_packets_total"),
		"number of undecrypted packets",
		[]string{"address", "stream_id"}, nil,
	)

	byteSentTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "sent_bytes_total"),
		"total number of sent data bytes, including retransmissions",
		[]string{"address", "stream_id"}, nil,
	)

	byteRecvTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_bytes_total"),
		"total number of received bytes",
		[]string{"address", "stream_id"}, nil,
	)

	byteRcvLossTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_lost_bytes_total"),
		"total number of lost bytes",
		[]string{"address", "stream_id"}, nil,
	)

	byteRetransTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "retransmitted_bytes_total"),
		"total number of retransmitted bytes",
		[]string{"address", "stream_id"}, nil,
	)

	byteSndDropTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "sent_dropped_bytes_total"),
		"number of too-late-to-send dropped bytes",
		[]string{"address", "stream_id"}, nil,
	)

	byteRcvDropTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_dropped_bytes_total"),
		"number of too-late-to play missing bytes (estimate based on average packet size)",
		[]string{"address", "stream_id"}, nil,
	)

	byteRcvUndecryptTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metrics.Namespace, srtSubsystem, "receive_undecrypted_bytes_total"),
		"number of undecrypted bytes",
		[]string{"address", "stream_id"}, nil,
	)
)

// Exporter collects metrics. It implements prometheus.Collector.
type Exporter struct {
	server *srt.Server
}

func NewExporter(s *srt.Server) *Exporter {
	e := Exporter{server: s}
	return &e
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- activeSocketsDesc
	ch <- pktSentTotalDesc
	ch <- pktRecvTotalDesc
	ch <- pktSndLossTotalDesc
	ch <- pktRcvLossTotalDesc
	ch <- pktRetransTotalDesc
	ch <- pktSentACKTotalDesc
	ch <- pktRecvACKTotalDesc
	ch <- pktSentNAKTotalDesc
	ch <- pktRecvNAKTotalDesc
	ch <- sndDurationTotalDesc
	ch <- pktSndDropTotalDesc
	ch <- pktRcvDropTotalDesc
	ch <- pktRcvUndecryptTotalDesc
	ch <- byteSentTotalDesc
	ch <- byteRecvTotalDesc
	ch <- byteRcvLossTotalDesc
	ch <- byteRetransTotalDesc
	ch <- byteSndDropTotalDesc
	ch <- byteRcvDropTotalDesc
	ch <- byteRcvUndecryptTotalDesc
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	stats := e.server.GetSocketStatistics()
	ch <- prometheus.MustNewConstMetric(activeSocketsDesc, prometheus.GaugeValue, float64(len(stats)))
	for _, stat := range stats {
		ch <- prometheus.MustNewConstMetric(pktSentTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktSent), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktRecvTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktRecv), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktSndLossTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktSendLoss), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktRcvLossTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktRecvLoss), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktRetransTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktRetrans), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktSentACKTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktSentACK), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktRecvACKTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktRecvACK), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktSentNAKTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktSentNAK), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktRecvNAKTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktRecvNAK), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(sndDurationTotalDesc, prometheus.CounterValue, float64(stat.Stats.UsSndDuration)/1_000_000.0, stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktSndDropTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktSendDrop), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktRcvDropTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktRecvDrop), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(pktRcvUndecryptTotalDesc, prometheus.CounterValue, float64(stat.Stats.PktRecvUndecrypt), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(byteSentTotalDesc, prometheus.CounterValue, float64(stat.Stats.ByteSent), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(byteRecvTotalDesc, prometheus.CounterValue, float64(stat.Stats.ByteRecv), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(byteRcvLossTotalDesc, prometheus.CounterValue, float64(stat.Stats.ByteRecvLoss), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(byteRetransTotalDesc, prometheus.CounterValue, float64(stat.Stats.ByteRetrans), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(byteSndDropTotalDesc, prometheus.CounterValue, float64(stat.Stats.ByteSendDrop), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(byteRcvDropTotalDesc, prometheus.CounterValue, float64(stat.Stats.ByteRecvDrop), stat.Address, stat.StreamID)
		ch <- prometheus.MustNewConstMetric(byteRcvUndecryptTotalDesc, prometheus.CounterValue, float64(stat.Stats.ByteRecvUndecrypt), stat.Address, stat.StreamID)
	}
}
