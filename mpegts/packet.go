package mpegts

import (
	"encoding/binary"
	"io"
)

// Packet represents a MPEGTS packet
type Packet struct {
	header          uint32
	payload         []byte
	adaptationField []byte
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
	ContinuityHdrMask = 0xf
)

/**
 * Packet Parsing
 */
// FromBytes parses a MPEG-TS packet from a byte slice
func (pkt *Packet) FromBytes(b []byte) error {
	if len(b) < PacketLen {
		return io.ErrUnexpectedEOF
	}

	if b[0] != SyncByte {
		return ErrInvalidPacket
	}

	pkt.header = binary.BigEndian.Uint32(b[0:HeaderLen])
	offset := HeaderLen

	// has adaptation field
	if pkt.header&AdaptationHdrMask > 0 {
		afLength := int(b[offset])
		if offset+afLength >= PacketLen {
			return ErrInvalidPacket
		}
		offset++
		pkt.adaptationField = b[offset : offset+afLength]
		offset += afLength
	} else {
		pkt.adaptationField = nil
	}

	// has payload
	if pkt.header&PayloadHdrMask > 0 {
		pkt.payload = b[offset:PacketLen]
	} else {
		pkt.payload = nil
	}

	return nil
}

// PID Payload ID
func (pkt *Packet) PID() uint16 {
	return uint16(pkt.header & PIDHdrMask >> PIDOffset)
}

// Continuity sequence number of payload packets
func (pkt *Packet) Continuity() byte {
	return byte(pkt.header & ContinuityHdrMask)
}

// PUSI the Payload Unit Start Indicator
func (pkt *Packet) PUSI() bool {
	return pkt.header&PUSIHdrMask > 0
}

func (pkt *Packet) Payload() []byte {
	return pkt.payload
}

func (pkt *Packet) AdaptationField() []byte {
	return pkt.adaptationField
}

/**
 * Packet creation
 */
// CreatePacket returns a bare MPEG-TS Packet
func CreatePacket(pid uint16) *Packet {
	var header uint32
	header |= uint32(SyncByte) << 24
	header |= uint32(pid&0x1fff) << 8
	return &Packet{
		header:          header,
		payload:         nil,
		adaptationField: nil,
	}
}

func (pkt *Packet) WithPUSI(pusi bool) *Packet {
	if pusi {
		pkt.header |= 0x1 << 22
	}
	return pkt
}

func (pkt *Packet) WithPayload(payload []byte) *Packet {
	pkt.payload = payload
	return pkt
}

func (pkt *Packet) WithAdaptationField(adaptationField []byte) *Packet {
	pkt.adaptationField = adaptationField
	return pkt
}

// ToBytes encodes a valid MPEGTS packet into a byte slice
// Expects a byte slice of atleast PacketLen
// Encoded packet is only valid if error is nil
func (pkt *Packet) ToBytes(data []byte) error {
	if len(data) < PacketLen {
		return io.ErrUnexpectedEOF
	}

	// Simplified MPEGTS packet header
	//   TEI always 0
	//   Transport priority always 0
	//   TSC always 0
	//   Continuity always 0
	if pkt.AdaptationField() != nil {
		pkt.header |= 0x1 << 5
	}
	if pkt.Payload() != nil {
		pkt.header |= 0x1 << 4
	}
	binary.BigEndian.PutUint32(data[0:4], pkt.header)
	offset := HeaderLen

	if pkt.AdaptationField() != nil {
		adaptationFieldLength := len(pkt.adaptationField)
		if adaptationFieldLength > PacketLen-offset-1 {
			return ErrDataTooLong
		}
		data[offset] = byte(adaptationFieldLength)
		offset++
		copy(data[offset:offset+adaptationFieldLength], pkt.adaptationField)
		offset += adaptationFieldLength
	}

	payloadLength := len(pkt.payload)
	if payloadLength > PacketLen-offset {
		return ErrDataTooLong
	}
	copy(data[offset:offset+payloadLength], pkt.payload)

	return nil
}

// Size returns the MPEGTS packet size
func (pkt Packet) Size() int {
	return PacketLen
}
