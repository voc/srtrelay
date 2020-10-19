package mpegts

import (
	"io"
	"log"
)

// PES constants
const (
	PESStartCode   = 0x000001
	MaxPayloadSize = PacketLen - 4
	PESMaxLength   = 200 * 1024

	PESStreamIDAudio = 0xc0
	PESStreamIDVideo = 0xe0
)

type ElementaryStream struct {
	parser    CodecParser
	pesWriter *io.PipeWriter
}

type CodecParser interface {
	HasInit() bool
	InitPacket() ([]byte, error)
	Parse(*io.PipeReader)
}

func encodePES(data []byte) ([]byte, error) {
	// Early check, but MPEG-TS packet will check again
	if len(data)+9+5 > PacketLen-HeaderLen {
		return nil, ErrDataTooLong
	}
	// len = len + 9 + 5;//pes len
	return nil, nil
}

// Parse PES packet
// assemble PES stream and call Codec parser
func (es *ElementaryStream) ParsePES(pkt *Packet) error {
	payload := pkt.Payload
	offset := 0

	// expect PES packet start code prefix
	if pkt.PUSI {
		// check start code
		if len(payload) < 6 || payload[0] != 0x0 || payload[1] != 0x0 || payload[2] != 0x1 {
			return ErrInvalidPacket
		}
		offset += 3

		// we are only interested in video codecs right now
		streamID := payload[offset]
		if streamID != PESStreamIDVideo {
			return ErrInvalidPacket
		}
		offset += 3

		if es.pesWriter != nil {
			// closes reader side
			err := es.pesWriter.Close()
			if err != nil {
				log.Println("got err from codec on close:", err)
			}
		}

		reader, writer := io.Pipe()
		es.pesWriter = writer

		// create new parser goroutine
		go es.parser.Parse(reader)
	}

	// Didn't yet receive a start packet
	if es.pesWriter == nil {
		return nil
	}

	// Feed bytes to parser synchronously
	_, err := es.pesWriter.Write(payload[offset:])
	if err != io.ErrClosedPipe {
		return err
	}
	return nil
}
