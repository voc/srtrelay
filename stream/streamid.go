package stream

import (
	"errors"
	"strings"

	"github.com/minio/minio/pkg/wildcard"
)

var (
	InvalidStreamID = errors.New("Invalid stream ID")
	InvalidMode     = errors.New("Invalid mode")
)

// Mode - client mode
type Mode uint8

const (
	_ Mode = iota
	ModePlay
	ModePublish
)

// StreamID represents a connection tuple
type StreamID struct {
	str      string
	mode     Mode
	name     string
	password string
}

// FromString reads a streamid from a string.
// The accepted stream id format is <mode>/<password>/<password>.
// The second slash and password is optional and defaults to empty.
// If error is not nil then StreamID will remain unchanged.
func (s *StreamID) FromString(src string) error {
	split := strings.Split(src, "/")

	password := ""
	if len(split) == 3 {
		password = split[2]
	} else if len(split) != 2 {
		return InvalidStreamID
	}
	modeStr := split[0]
	name := split[1]

	if len(name) == 0 {
		return InvalidStreamID
	}

	var mode Mode
	switch modeStr {
	case "play":
		mode = ModePlay
	case "publish":
		mode = ModePublish
	default:
		return InvalidMode
	}

	s.str = src
	s.mode = mode
	s.name = name
	s.password = password
	return nil
}

// Match checks a streamid against a string with wildcards.
// The string may contain * to match any number of characters.
func (s StreamID) Match(pattern string) bool {
	return wildcard.MatchSimple(pattern, s.str)
}

func (s StreamID) String() string {
	return s.str
}

func (s StreamID) Mode() Mode {
	return s.mode
}

func (s StreamID) Name() string {
	return s.name
}

func (s StreamID) Password() string {
	return s.password
}
