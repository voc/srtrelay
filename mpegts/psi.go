package mpegts

import (
	"encoding/binary"
	"io"
)

// PSI constants
const (
	PSIHeaderLen = 8
)

// PSI table type constants
const (
	TableTypePAT = 0x0
	TableTypePMT = 0x2
)

// Elementary Stream StreamType constants
const (
	StreamTypeAudio    = 0xf
	StreamTypeAVCVideo = 0x1b
)

// PSIHeader struct
type PSIHeader struct {
	tableID           byte
	sectionLength     uint16
	versionNumber     byte
	currentNext       bool // true means current table version is valid, false means current table version not yet valid
	sectionNumber     byte // number of current section
	lastSectionNumber byte // number of last table section
}

// ParsePSIHeader func
func ParsePSIHeader(data []byte) (*PSIHeader, error) {
	hdr := PSIHeader{}
	if len(data) < 3 {
		return nil, io.ErrUnexpectedEOF
	}
	hdr.tableID = data[0]
	hdr.sectionLength = binary.BigEndian.Uint16(data[1:3]) & 0xfff

	if len(data) < int(3+hdr.sectionLength) {
		return nil, io.ErrUnexpectedEOF
	}

	hdr.versionNumber = data[5] >> 1 & 0x1f
	currentNext := data[5] & 0x1
	if currentNext == 1 {
		hdr.currentNext = true
	}

	hdr.sectionNumber = data[6]
	hdr.lastSectionNumber = data[7]
	// log.Println("PSI: tableID", hdr.tableID, "sectionLength", hdr.sectionLength, "version", hdr.versionNumber,
	// "currentNext", hdr.currentNext, "sectionNumber", hdr.sectionNumber, "lastSectionNumber", hdr.lastSectionNumber)

	return &hdr, nil
}
