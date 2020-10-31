package mpegts

import (
	"encoding/binary"
	"io"
	"log"
)

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
	parser    CodecParser
	pesWriter *io.PipeWriter
}

type CodecParser interface {
	HasInit() bool
	InitPacket() ([]byte, error)
	Parse(*io.PipeReader)
}

// Encode video codec PES with custom payload
func encodeVideoPES(data []byte) ([]byte, error) {
	// Early check, but MPEG-TS packet will check again
	if len(data)+PESHeaderSize > PacketLen-HeaderLen {
		return nil, ErrDataTooLong
	}

	pes := make([]byte, PESHeaderSize, MaxPayloadSize)

	// write pes header (Video only for now)
	offset := 0
	tmp := uint32(PESStartCode<<8) | PESStreamIDVideo
	binary.BigEndian.PutUint32(pes[offset:offset+4], tmp)
	offset += 4

	// write pes length (can be 0 only for Video packets)
	binary.BigEndian.PutUint16(pes[offset:offset+2], 0)

	// magic?! flag + pts
	// pes = append(pes, 0x80, 0x80, 5, 0, 0, 0, 0, 0)

	pes = append(pes, data...)

	return pes, nil
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
