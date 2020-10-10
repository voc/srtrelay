package mpegts

import (
	"encoding/binary"
	"io"
	"log"
)

const (
	Magic = 0x47

	HeaderLen = 4
	PacketLen = 188
	PIDOffset = 8

	PUSIHdrMask       = 0x400000
	PIDHdrMask        = 0x1fff00
	AdaptationHdrMask = 0x20
	PayloadHdrMask    = 0x10
)

// PIDs
const (
	PIDPAT  = 0x0    // Program Association Table (PAT) contains a directory listing of all Program Map Tables.
	PIDCAT  = 0x1    // Conditional Access Table (CAT) contains a directory listing of all ITU-T Rec. H.222 entitlement management message streams used by Program Map Tables.
	PIDTSDT = 0x2    // Transport Stream Description Table (TSDT) contains descriptors related to the overall transport stream
	PIDNull = 0x1fff // Null Packet (used for fixed bandwidth padding)
)

type Parser struct {
	hasInit bool
	init    [][]byte
	pm      map[uint16]uint16 // map[ProgramMapID]ProgramNumber
}

// 1. Parse PAT to get PID->PMT mappings
// 2. Parse PMTs to Find PID->PES mappings
// 3. Parse PES to find H.264 SPS or equivalent
// Store the whole shebang in order and send it to the client
// Maybe remember PAT+PMTs for a whole stream?

func (p *Parser) Parse(data []byte) error {
	for {
		if len(data) == 0 {
			return nil
		}
		pkt := Packet{}
		err := pkt.FromBytes(data)
		if err != nil {
			// Incomplete packet, TODO: keep rest data
			if err == io.ErrUnexpectedEOF {
				log.Fatalln("parsing failed, incomplete packet")
				return nil
			}
			return err
		}

		// parse table
		if pkt.PID == PIDPAT && pkt.PUSI {
			p.ParsePSI(pkt.Payload)

		}

		data = data[pkt.Size():]
	}
}

const (
	TableTypePAT = 0x0
	TableTypePMT = 0x2
)

func (p *Parser) ParsePSI(payload []byte) error {
	ptr := int(payload[0])
	log.Println("table ptr", ptr)
	offset := 1 + ptr

	if len(payload)-offset < 3 {
		return io.ErrUnexpectedEOF
	}
	tid := payload[offset]
	offset++
	sectionLen := binary.BigEndian.Uint16(payload[offset : offset+2])
	log.Println("tid", tid, "sectionLength", sectionLen)
	return nil
}

// isPSIPayload checks whether the payload is a PSI one
func (p *Parser) PSIPID(pid uint16) bool {
	_, knownPID := p.pm[pid]
	return pid == PIDPAT || // PAT
		knownPID || // PMT
		((pid >= 0x10 && pid <= 0x14) || (pid >= 0x1e && pid <= 0x1f)) //DVB
}

// Return data needed for decoder init or nil
func (p *Parser) InitData() [][]byte {
	if !p.hasInit {
		return nil
	}

	return p.init
}
