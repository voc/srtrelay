package mpegts

// AVC NAL unit type constants
const (
	HEVCNALUnitTypeVPS = 32
	HEVCNALUnitTypeSPS = 33
	HEVCNALUnitTypePPS = 34
)

// H265Parser parser for h.265 (HEVC) init packets
type H265Parser struct{}

// ContainsInit checks whether the MPEG-TS packet contains a H.265 VPS, SPS, or PPS
func (p H265Parser) ContainsInit(pkt *Packet) (bool, error) {
	var state byte
	buf := pkt.Payload()
	for i := 0; i < len(buf); i++ {
		if state == 0x57 {
			// H.265 NAL unit type is in bits 1–6 of the first byte after start code
			nalType := (buf[i] >> 1) & 0x3F
			switch nalType {
			case HEVCNALUnitTypeVPS,
				HEVCNALUnitTypeSPS,
				HEVCNALUnitTypePPS:
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
