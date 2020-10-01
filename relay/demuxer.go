package relay

import (
	"github.com/voc/srtrelay/mpegts"
)

type TransportType uint

const (
	Unknown = 0
	MpegTs  = 1
)

// Demuxer used for finding synchronization point for onboarding a client
type Demuxer struct {
	transport TransportType
	parser    mpegts.Parser
}

func DetermineTransport(data []byte) TransportType {
	// Detect transport type
	if len(data) >= mpegts.HeaderLen && data[0] == mpegts.Magic {
		return MpegTs
	}

	return Unknown
}

func (d *Demuxer) FindInit(data []byte) ([][]byte, error) {

	if d.transport == Unknown {
		d.transport = DetermineTransport(data)
	}

	switch d.transport {
	case MpegTs:
		err := d.parser.Parse(data)
		if err != nil {
			return nil, err
		}
		res := d.parser.InitData()
		return res, nil
	default:
		return make([][]byte, 0), nil
	}
}
