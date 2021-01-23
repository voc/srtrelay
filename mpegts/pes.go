package mpegts

// PES constants
const (
	PESStartCode   = 0x000001
	MaxPayloadSize = PacketLen - 4
	PESHeaderSize  = 6
	PESMaxLength   = 200 * 1024

	PESStreamIDAudio = 0xc0
	PESStreamIDVideo = 0xe0
)

type ElementaryStream struct {
	HasInit     bool
	CodecParser CodecParser
}

type CodecParser interface {
	ContainsInit(pkt *Packet) (bool, error)
}
