package mpegts

// AVC NAL unit type constants
const (
	NALUnitCodedSliceIDR = 5
	NALUnitTypeSPS       = 7
	NALUnitTypePPS       = 8
)

// H264Parser parser for h.264 init packets
type H264Parser struct{}

// ContainsInit checks whether the MPEG-TS packet contains a h.264 PPS or SPS
func (p H264Parser) ContainsInit(pkt *Packet) (bool, error) {
	var state byte
	buf := pkt.Payload()
	for i := 0; i < len(buf); i++ {
		if state == 0x57 {
			nalType := buf[i] & 0x1F
			switch nalType {
			case NALUnitTypeSPS:
				fallthrough
			case NALUnitTypePPS:
				return true, nil
			}
		}

		cur := 0
		switch buf[i] {
		case 0x00:
			cur = 1
		case 0x01:
			cur = 3
		default:
			cur = 2
		}

		/* state of last four bytes packed into one byte; two bits for unseen/zero/over
		 * one/one (0..3 respectively).
		 */
		state = (state << 2) | byte(cur)
	}
	return false, nil
}
