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

// StreamType constants
const (
	StreamTypeH264 = 0x1b
	StreamTypeH265 = 0x24
)

// Parser object for finding the synchronization point in a MPEGTS stream
// This works as follows: parse packets until all the following have been fulfilled
//   1. Parse PAT to get PID->PMT mappings
//   2. Parse PMTs to Find PID->PES mappings
//   3. Parse PES to find H.264 SPS+PPS or equivalent
// Store the original packets or generate new ones to send to a client
type Parser struct {
	init               [][]byte                     // collected packets to initialize a decoder
	expectedPATSection byte                         // id of next expected PAT section
	expectedPMTSection byte                         // id of next expected PMT sectino
	hasPAT             bool                         // MPEG-TS PAT packet stored
	hasPMT             bool                         // MPEG-TS PMT packet stored
	pmtMap             map[uint16]uint16            // map[pid]programNumber
	tspMap             map[uint16]*ElementaryStream // transport stream program map
}

func NewParser() *Parser {
	return &Parser{
		pmtMap: make(map[uint16]uint16),
		tspMap: make(map[uint16]*ElementaryStream),
		init:   make([][]byte, 0, 3),
	}
}

func (p *Parser) hasInit() bool {
	if !p.hasPAT || !p.hasPMT {
		return false
	}

	for _, stream := range p.tspMap {
		if !stream.parser.HasInit() {
			return false
		}
	}

	return true
}

// InitData returns data needed for decoder init or nil if the parser is not ready yet
func (p *Parser) InitData() ([][]byte, error) {
	if !p.hasInit() {
		return nil, nil
	}

	// add stream init packets
	for _, stream := range p.tspMap {
		packet, err := stream.parser.InitPacket()
		if err != nil {
			return nil, err
		}
		p.init = append(p.init, packet)
	}

	return p.init, nil
}

// Parse processes all MPEGTS packets from a buffer
func (p *Parser) Parse(data []byte) error {
	pkt := Packet{}
	for {
		if len(data) == 0 {
			return nil
		}
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

		} else if stream, ok := p.tspMap[pkt.PID]; ok {
			// Parse PES and store SPS+PPS packet
			// log.Println("payload", len(pkt.Payload), len(pkt.AdaptationField))
			err = stream.ParsePES(&pkt)
			if err != nil {
				return err
			}
		}

		// store init packets and all remaining packets from the buffer after init
		if storePacket || p.hasInit() {
			p.init = append(p.init, data[:pkt.Size()])
		}

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

			switch streamType {
			case StreamTypeH264:
				p.tspMap[elementaryPID] = &ElementaryStream{
					parser: NewH264Parser(elementaryPID),
				}
			}

			if offset >= end {
				break
			}
		}

		shouldStore = true
		p.expectedPMTSection = hdr.sectionNumber + 1
		if hdr.sectionNumber == hdr.lastSectionNumber {
			p.hasPMT = true
		}
	}
	return shouldStore, nil
}
