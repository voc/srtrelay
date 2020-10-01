package mpegts

import (
	"encoding/binary"
	"io"
	"log"
)

type Packet struct {
	PID             uint16 // Payload ID
	PUSI            bool   // Payload Unit Start Indicator
	Payload         []byte
	AdaptationField []byte
}

func (pkt Packet) Size() int {
	return PacketLen
}

func (pkt *Packet) FromBytes(b []byte) error {
	if len(b) < PacketLen {
		log.Println("pkt too short")
		return io.ErrUnexpectedEOF
	}

	if b[0] != Magic {
		log.Fatalln("parser error")
	}

	hdr := binary.BigEndian.Uint32(b[0:HeaderLen])
	pkt.PID = uint16(hdr & PIDHdrMask >> PIDOffset)
	pkt.PUSI = hdr&PUSIHdrMask > 0
	// log.Println("PID", pkt.Pid, "PUSI", Pusi)

	// Parse adaptation field
	offset := HeaderLen
	if hdr&AdaptationHdrMask > 0 {
		afLength := int(b[offset])
		pkt.AdaptationField = b[offset : offset+afLength+1]
		offset += afLength + 1
		// log.Println("aflength", afLength)
	}

	// parse payload
	if hdr&PayloadHdrMask > 0 {
		pkt.Payload = b[offset:PacketLen]
	}

	return nil
}

func (pkt *Packet) ParsePayload() {

}
