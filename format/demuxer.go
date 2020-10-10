package format

import (
	"github.com/voc/srtrelay/mpegts"
)

// TransportType type
type TransportType uint

// TransportType constants
const (
	Unknown = 0
	MpegTs  = 1
)

// Demuxer used for finding synchronization point for onboarding a client
type Demuxer struct {
	transport TransportType
	parser    mpegts.Parser
}

// DetermineTransport tries to detect the type of transport from the stream
// If the type is not clear it returns Unknown
func DetermineTransport(data []byte) TransportType {
	if len(data) >= mpegts.HeaderLen && data[0] == mpegts.Magic {
		return MpegTs
	}

	return Unknown
}

// FindInit determines a synchronization point in the stream
// Finally it returns the required stream packets up to that point
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
