package mpegts

import (
	"encoding/binary"
	"io"
)

// Packet represents a MPEGTS packet
type Packet struct {
	PID             uint16 // Payload ID
	PUSI            bool   // Payload Unit Start Indicator
	Payload         []byte
	AdaptationField []byte
}

// MPEGTS Packet constants
const (
	SyncByte = 0x47

	HeaderLen = 4
	PacketLen = 188
	PIDOffset = 8

	PUSIHdrMask       = 0x400000
	PIDHdrMask        = 0x1fff00
	AdaptationHdrMask = 0x20
	PayloadHdrMask    = 0x10
)

// FromBytes creates a packet from a byte slice
func (pkt *Packet) FromBytes(b []byte) error {
	if len(b) < PacketLen {
		return io.ErrUnexpectedEOF
	}

	if b[0] != SyncByte {
		return ErrInvalidPacket
	}

	hdr := binary.BigEndian.Uint32(b[0:HeaderLen])
	pkt.PID = uint16(hdr & PIDHdrMask >> PIDOffset)
	pkt.PUSI = hdr&PUSIHdrMask > 0
	offset := HeaderLen

	// has adaptation field
	if hdr&AdaptationHdrMask > 0 {
		afLength := int(b[offset])
		pkt.AdaptationField = b[offset : offset+afLength+1]
		offset += afLength + 1
	} else {
		pkt.AdaptationField = nil
	}

	// has payload
	if hdr&PayloadHdrMask > 0 {
		pkt.Payload = b[offset:PacketLen]
	} else {
		pkt.Payload = nil
	}

	return nil
}

// ToBytes encodes a valid MPEGTS packet into a byte slice
// Expects a byte slice of atleast PacketLen
// Encoded packet is only valid if error is nil
func (pkt *Packet) ToBytes(data []byte) error {
	if len(data) < PacketLen {
		return io.ErrUnexpectedEOF
	}
	data[0] = SyncByte

	// Simplified MPEGTS packet header
	//   TEI always 0
	//   Transport priority always 0
	//   TSC always 0
	//   Continuity always 0
	var hdr uint32
	hdr |= uint32(pkt.PID&0x1fff) << 8
	if pkt.AdaptationField != nil {
		hdr |= 0x1 << 5
	}
	if pkt.Payload != nil {
		hdr |= 0x1 << 4
	}
	binary.BigEndian.PutUint32(data[0:4], hdr)

	offset := HeaderLen
	adaptationFieldLength := len(pkt.AdaptationField)
	if adaptationFieldLength > PacketLen-offset {
		return ErrDataTooLong
	}
	copy(data[offset:offset+adaptationFieldLength], pkt.AdaptationField)
	offset += adaptationFieldLength

	payloadLength := len(pkt.Payload)
	if payloadLength > PacketLen-offset {
		return ErrDataTooLong
	}
	copy(data[offset:offset+payloadLength], pkt.Payload)
	offset += payloadLength

	return nil
}

// Size returns the MPEGTS packet size
func (pkt Packet) Size() int {
	return PacketLen
}
