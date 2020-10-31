package mpegts

import (
	"bufio"
	"io"
	"log"
)

// AVC NAL constants
const (
	NALStartCode  = 0x00000001 // NAL Start code with single zero byte, required for SPS, PPS and first NAL per picture
	NALHeaderSize = 4
)

// AVC NAL unit type constants
const (
	NALUnitTypeSPS = 7
	NALUnitTypePPS = 8
)

// H264Parser struct
type H264Parser struct {
	pid        uint16        // stream PID
	sps        []byte        // SPS NAL
	pps        []byte        // PPS NAL
	initPacket []byte        // packet for init
	done       chan struct{} // success signal channel
	semaphor   chan struct{} // semaphor channel
}

// NewH264Parser creates a new H.264 elementary stream parser
// Parse can safely be run in a separate goroutine from HasInit and InitPacket
func NewH264Parser(pid uint16) *H264Parser {
	semaphor := make(chan struct{}, 1)
	semaphor <- struct{}{}
	return &H264Parser{
		pid:      pid,
		done:     make(chan struct{}),
		semaphor: semaphor,
	}
}

/**
 * CodecParser Implementation
 */
// HasInit returns true if the ES has gathered all data required for init
func (h *H264Parser) HasInit() bool {
	select {
	case <-h.done:
		return true
	default:
		return false
	}
}

// InitPacket creates a MPEG2-TS PES-Packet containing H.264 SPS and PPS
func (h *H264Parser) InitPacket() ([]byte, error) {
	// put SPS and PPS into pes payload
	pesDataLen := len(h.sps) + len(h.pps)
	pesData := make([]byte, pesDataLen)
	copy(pesData[:len(h.sps)], h.sps)
	copy(pesData[len(h.sps):], h.pps)

	// encode PES payload
	pesPayload, err := encodeVideoPES(pesData)
	if err != nil {
		return nil, err
	}

	// pad with adaptationField
	adaptationLen := MaxPayloadSize - len(pesPayload)
	adaptationField := make([]byte, adaptationLen)
	adaptationField[0] = 0x3 << 6
	for i := 0; i < adaptationLen-1; i++ {
		adaptationField[i+1] = 0xff
	}

	// create MPEG-TS Packet
	pkt := Packet{
		PID:             h.pid,
		PUSI:            true,
		Payload:         pesPayload,
		AdaptationField: adaptationField,
	}

	data := make([]byte, PacketLen)
	err = pkt.ToBytes(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Parse reads H.264 PPS and SPS from PES payloads
func (h *H264Parser) Parse(rd *io.PipeReader) {
	// get semaphor
	<-h.semaphor
	defer func() { h.semaphor <- struct{}{} }()

	// Don't read after success
	if h.HasInit() {
		rd.Close()
		return
	}

	var previousNalType byte
	var nalBuffer []byte
	skippedZero := false

	// buffered reader for reading cross-packet data
	brd := bufio.NewReaderSize(rd, MaxPayloadSize)

	for {
		// Peek start of potential NAL packet
		nalStart, err := brd.Peek(NALHeaderSize)
		if err != nil {
			break
		}

		// Check starting bytes
		if nalStart[0] == 0x0 && nalStart[1] == 0x0 && nalStart[2] == 0x1 ||
			skippedZero && nalStart[0] == 0x0 && nalStart[1] == 0x1 {

			typeOffset := 3
			if skippedZero {
				typeOffset = 2
			}

			// parse nal type
			nalType := nalStart[typeOffset] & 0x1f
			// log.Println("NAL", nalType)

			forbiddenZero := nalStart[typeOffset] >> 7 & 1
			if forbiddenZero != 0 {
				log.Println("forbidden zero wasn't zero")
				continue
			}

			// store previous SPS/PPS
			switch previousNalType {
			case NALUnitTypeSPS:
				h.sps = nalBuffer
				nalBuffer = nil
				if h.sps != nil && h.pps != nil {
					close(h.done)
					rd.Close()
					return
				}

			case NALUnitTypePPS:
				h.pps = nalBuffer
				nalBuffer = nil
				if h.sps != nil && h.pps != nil {
					close(h.done)
					rd.Close()
					return
				}
			}

			// log.Printf("nal, offset: %x, unit type: %x\n", offset, nalType)
			if nalType == NALUnitTypeSPS || nalType == NALUnitTypePPS {
				// log.Println("got SPS/PPS")
				nalBuffer = make([]byte, 0, 200)
				nalBuffer = append(nalBuffer, 0x0)

				if skippedZero {
					nalBuffer = append(nalBuffer, 0x0)
				}

				// ignore error, because the bytes should be buffered through peek
				startSlice := make([]byte, NALHeaderSize)
				n, _ := brd.Read(startSlice)
				if n < NALHeaderSize {
					log.Fatal("Short read, should be buffered")
					return
				}
				nalBuffer = append(nalBuffer, startSlice...)
			}
			previousNalType = nalType
		}

		// Read until next zero byte
		skippedZero = false
		buf, err := brd.ReadSlice(0x0)
		if nalBuffer != nil {
			// may append an extra zero byte at the end of the NAL
			// which is okay
			nalBuffer = append(nalBuffer, buf...)
		}

		// brd buffer full, continue parsing
		if err == bufio.ErrBufferFull {
			continue
		}

		if err != nil {
			break
		}

		// err is nil so we skipped a zero byte
		skippedZero = true
	}

	// read remaining bytes
	buffered := brd.Buffered()
	if buffered > 0 {
		tmp := make([]byte, buffered)
		n, _ := brd.Read(tmp)
		nalBuffer = append(nalBuffer, tmp[:n]...)
	}

	// store previous SPS/PPS
	// assumes last nal continues until the end
	switch previousNalType {
	case NALUnitTypeSPS:
		h.sps = nalBuffer
	case NALUnitTypePPS:
		h.pps = nalBuffer
	}

	if h.sps != nil && h.pps != nil {
		close(h.done)
	}
	rd.Close()

	// Parse PTS
	// log.Printf("got ES start, stream id: %x, length: %d, pid: %d\n", streamID, packetLength, pid)
}
