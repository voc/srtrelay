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

	// create MPEG-TS Packet
	pkt := Packet{
		PID:     h.pid,
		PUSI:    true,
		Payload: pesPayload,
	}

	// //sps pps
	// int len = ti->sps_len + ti->pps_len;
	// if (len > TS_PACK_LEN-4) {
	//     printf("pid=%d, pes size=%d is abnormal!!!!\n", pid, len);
	//     return ret;
	// }
	// pos ++;
	// //pid
	// ti->es_pid = pid;
	// tmp = ti->es_pid >> 8;
	// p[pos++] = 0x40 | tmp;
	// tmp = ti->es_pid;
	// p[pos++] = tmp;
	// p[pos] = 0x10;
	// int ad_len = TS_PACK_LEN - 4 - len - 1;
	// if (ad_len > 0) {
	//     p[pos++] = 0x30;
	//     p[pos++] = ad_len;//adaptation length
	//     p[pos++] = 0x00;//
	//     memset(p + pos, 0xFF, ad_len-1);
	//     pos += ad_len - 1;
	// }else{
	//     pos ++;
	// }

	// //pes
	// p[pos++] = 0;
	// p[pos++] = 0;
	// p[pos++] = 1;
	// p[pos++] = stream_id;
	// p[pos++] = 0;//total size
	// p[pos++] = 0;//total size
	// p[pos++] = 0x80;//flag
	// p[pos++] = 0x80;//flag
	// p[pos++] = 5;//header_len
	// p[pos++] = 0;//pts
	// p[pos++] = 0;
	// p[pos++] = 0;
	// p[pos++] = 0;
	// p[pos++] = 0;
	// memcpy(p+pos, ti->sps, ti->sps_len);
	// pos += ti->sps_len;
	// memcpy(p+pos, ti->pps, ti->pps_len);
	// pos += ti->pps_len;

	data := make([]byte, PacketLen)
	err = pkt.ToBytes(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// func readByte(rd io.Reader) (byte, error) {
// 	b := make([]byte, 1)
// 	_, err := rd.Read(b)
// 	return b[0], err
// }

// var counter = 0

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
				log.Println("got SPS len", len(h.sps))
				if h.sps != nil && h.pps != nil {
					close(h.done)
					rd.Close()
					return
				}

			case NALUnitTypePPS:
				h.pps = nalBuffer
				nalBuffer = nil
				log.Println("got PPS len", len(h.pps))
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

				startSlice := nalBuffer[:NALHeaderSize]
				if skippedZero {
					nalBuffer = append(nalBuffer, 0x0)
					startSlice = nalBuffer[1 : NALHeaderSize+1]
				}

				// ignore error, because the bytes should be buffered through peek
				n, _ := brd.Read(startSlice)
				if n < NALHeaderSize {
					log.Fatal("Short read, should be buffered")
					return
				}
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
		log.Println("got SPS len on end", len(h.sps))
	case NALUnitTypePPS:
		h.pps = nalBuffer
		log.Println("got PPS len on end", len(h.pps))
	}

	if h.sps != nil && h.pps != nil {
		close(h.done)
	}
	rd.Close()

	// Parse PTS
	// log.Printf("got ES start, stream id: %x, length: %d, pid: %d\n", streamID, packetLength, pid)
}
