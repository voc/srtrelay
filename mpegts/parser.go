package mpegts

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
)

// MPEGTS errors
var (
	ErrDataTooLong   = errors.New("Data too long")
	ErrInvalidPacket = errors.New("Invalid MPEGTS packet")
)

// PID constants
const (
	PIDPAT  = 0x0    // Program Association Table (PAT) contains a directory listing of all Program Map Tables.
	PIDTSDT = 0x2    // Transport Stream Description Table (TSDT) contains descriptors related to the overall transport stream
	PIDNull = 0x1fff // Null Packet (used for fixed bandwidth padding)
)

// Parser object for finding the synchronization point in a MPEGTS stream
// This works as follows: parse packets until all the following have been fulfilled
//   1. Parse PAT to get PID->PMT mappings
//   2. Parse PMTs to Find PID->PES mappings
//   3. Parse PES to find H.264 SPS+PPS or equivalent
// Store the original packets or generate new ones to send to a client
type Parser struct {
	init               [][]byte
	expectedPATSection byte
	expectedPMTSection byte
	hasPAT             bool
	hasPMT             bool
	hasInit            bool
	hasSPS             bool
	hasPPS             bool
	pmtMap             map[uint16]uint16 // map[pid]programNumber
	esMap              map[uint16]byte   // map[pid]streamType
}

func NewParser() *Parser {
	return &Parser{
		pmtMap: make(map[uint16]uint16),
		esMap:  make(map[uint16]byte),
	}
}

// InitData returns data needed for decoder init or nil if the parser is not ready yet
func (p *Parser) InitData() [][]byte {
	if !p.hasPAT || !p.hasPMT || !p.hasSPS || !p.hasPPS {
		return nil
	}

	return p.init
}

// Parse processes all MPEGTS packets from a buffer
func (p *Parser) Parse(data []byte) error {
	for {
		if len(data) == 0 {
			return nil
		}
		pkt := Packet{}
		err := pkt.FromBytes(data)
		if err != nil {
			// Incomplete packet, TODO: keep rest data?
			if err == io.ErrUnexpectedEOF {
				log.Fatalln("parsing failed, incomplete packet")
				return nil
			}
			return err
		}

		storePacket := false
		if pkt.PID == PIDPAT {
			// parse PMT and store PAT packet
			storePacket, err = p.ParsePSI(pkt.Payload)
			if err != nil {
				return err
			}

		} else if _, ok := p.pmtMap[pkt.PID]; ok {
			// Parse PID->ES mapping and store PMT packet
			storePacket, err = p.ParsePSI(pkt.Payload)
			if err != nil {
				return err
			}

		} else if _, ok := p.esMap[pkt.PID]; ok {
			// Parse PES and store SPS+PPS packet
			storePacket, err = p.ParsePES(pkt.Payload, pkt.PID)
			if err != nil {
				return err
			}
		}

		if storePacket {
			p.init = append(p.init, data[:pkt.Size()])
		}

		// parse PMT and store

		data = data[pkt.Size():]
	}
}

// ParsePSI selectively parses a Program Specific Information (PSI) table
// We are only interested in PAT and PMT
func (p *Parser) ParsePSI(data []byte) (bool, error) {
	// skip to section header
	ptr := int(data[0])
	offset := 1 + ptr
	shouldStore := false

	hdr, err := ParsePSIHeader(data[offset:])
	if err != nil {
		return false, err
	}
	offset += PSIHeaderLen

	// We are only interested in PAT and PMT
	switch hdr.tableID {
	case TableTypePAT:
		// expect program map in order
		if p.expectedPATSection != hdr.sectionNumber || !hdr.currentNext {
			return false, nil
		}

		end := offset + int(hdr.sectionLength-9)/4
		for {
			programNumber := binary.BigEndian.Uint16(data[offset : offset+2])
			offset += 2
			pid := binary.BigEndian.Uint16(data[offset:offset+2]) & 0x1fff
			offset += 2
			if programNumber != 0 {
				p.pmtMap[pid] = programNumber
			}
			if offset >= end {
				break
			}
		}
		shouldStore = true
		p.expectedPATSection = hdr.sectionNumber + 1
		if hdr.sectionNumber == hdr.lastSectionNumber {
			log.Println("got PAT")
			p.hasPAT = true
		}

	case TableTypePMT:
		// expect program map in order
		if p.expectedPMTSection != hdr.sectionNumber || !hdr.currentNext {
			return false, nil
		}

		// skip PCR PID
		offset += 2

		programInfoLength := binary.BigEndian.Uint16(data[offset:offset+2]) & 0xfff
		offset += 2 + int(programInfoLength)

		end := offset + int(hdr.sectionLength-programInfoLength-13)
		for {
			streamType := data[offset]
			offset++

			elementaryPID := binary.BigEndian.Uint16(data[offset:offset+2]) & 0x1fff
			offset += 2

			esInfoLength := binary.BigEndian.Uint16(data[offset:offset+2]) & 0xfff
			offset += 2 + int(esInfoLength)

			// log.Println("stream type", streamType, "elementary pid", elementaryPID, "esInfoLength", esInfoLength)
			p.esMap[elementaryPID] = streamType

			if offset >= end {
				break
			}
		}

		shouldStore = true
		p.expectedPMTSection = hdr.sectionNumber + 1
		if hdr.sectionNumber == hdr.lastSectionNumber {
			log.Println("got PMT")
			p.hasPMT = true
		}
	}
	return shouldStore, nil
}

// PES constants
const (
	PESStartCode   = 0x000001
	MaxPayloadSize = PacketLen - 4
	PESMaxLength   = 200 * 1024

	PESStreamIDAudio = 0xc0
	PESStreamIDVideo = 0xe0
)

// AVC NAL constants
const (
	NALStartCode  = 0x00000001 // NAL Start code with single zero byte, required for SPS, PPS and first NAL per picture
	NALHeaderSize = 6
)

// AVC NAL unit type constants
const (
	NALUnitTypeSPS = 7
	NALUnitTypePPS = 8
)

func (p *Parser) ParsePES(data []byte, pid uint16) (bool, error) {
	var shouldStore bool

	if !isPESPayload(data) {
		return false, nil
	}

	// start := binary.BigEndian.Uint32(data[0:4])
	// packetStartCode := start >> 8 & 0xffffff
	if data[0] != 0x0 || data[1] != 0x0 || data[2] != 0x1 {
		// log.Println("invalid PES start code")
		return false, nil
		// return false, ErrInvalidPacket
	}
	offset := 3

	streamID := data[offset]
	if streamID != PESStreamIDVideo {
		return false, nil
	}
	offset++

	packetLength := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	// unbounded packet, only allowed for video ES
	if packetLength == 0 {
		packetLength = PESMaxLength
	}
	offset += 2

	// Just look for NAL units...
	for {
		if offset > len(data)-NALHeaderSize {
			break
		}

		// Find NAL
		// start := binary.BigEndian.Uint32(data[offset+1:offset+5]) >> 8
		if data[offset] == 0x0 && data[offset+1] == 0x0 && data[offset+2] == 0x0 && data[offset+3] == 0x1 {
			// log.Printf("nal header: 0x%x", data[offset:offset+6])

			offset += 4

			forbiddenZero := data[offset] >> 7 & 1
			if forbiddenZero != 0 {
				offset++
				// log.Println("forbidden zero wasn't zero")
				continue
			}

			// refIdc := data[offset] >> 5 & 0x3

			unitType := data[offset] & 0x1f
			// log.Printf("nal, offset: %x, unit type: %x, ref idc: %x\n", offset, unitType, refIdc)
			if unitType == NALUnitTypeSPS {
				log.Println("got SPS")
				shouldStore = true
				p.hasSPS = true
			} else if unitType == NALUnitTypePPS {
				log.Println("got PPS")
				shouldStore = true
				p.hasPPS = true
			}
		}
		offset++
	}

	// Parse PTS
	// log.Printf("got ES start, stream id: %x, length: %d, pid: %d\n", streamID, packetLength, pid)
	return shouldStore, nil
}
