package stream

import (
	"errors"
	"fmt"
	"strings"

	"github.com/minio/minio/pkg/wildcard"
)

var (
	InvalidSlashes      = errors.New("Invalid number of slashes, must be 1 or 2")
	InvalidMode         = errors.New("Invalid mode")
	MissingName         = errors.New("Missing name after slash")
	InvalidNamePassword = errors.New("Name/Password is not allowed to contain slashes")
)

// Mode - client mode
type Mode uint8

const (
	_ Mode = iota
	ModePlay
	ModePublish
)

func (m Mode) String() string {
	switch m {
	case ModePlay:
		return "play"
	case ModePublish:
		return "publish"
	default:
		return "unknown"
	}
}

// StreamID represents a connection tuple
type StreamID struct {
	str      string
	mode     Mode
	name     string
	password string
}

// Creates new StreamID
// returns error if mode is invalid.
// id is nil on error
func NewStreamID(name string, password string, mode Mode) (*StreamID, error) {
	id := &StreamID{
		name:     name,
		password: password,
		mode:     mode,
	}
	var err error
	id.str, err = id.toString()
	if err != nil {
		return nil, err
	}
	return id, nil
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
		return InvalidSlashes
	}
	modeStr := split[0]
	name := split[1]

	if len(name) == 0 {
		return MissingName
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

// toString returns a string representation of the streamid
func (s *StreamID) toString() (string, error) {
	mode := ""
	if s.mode == ModePlay {
		mode = "play"
	} else if s.mode == ModePublish {
		mode = "publish"
	} else {
		return "", InvalidMode
	}
	if strings.Contains(s.name, "/") {
		return "", InvalidNamePassword
	}
	if strings.Contains(s.password, "/") {
		return "", InvalidNamePassword
	}
	if len(s.password) == 0 {
		return fmt.Sprintf("%s/%s", mode, s.name), nil
	}
	return fmt.Sprintf("%s/%s/%s", mode, s.name, s.password), nil
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
