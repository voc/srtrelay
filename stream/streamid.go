package stream

import (
	"errors"
	"fmt"
	"strings"

	"github.com/IGLOU-EU/go-wildcard/v2"
)

const IDPrefix = "#!::"

var (
	ErrInvalidSlashes      = errors.New("invalid number of slashes, must be 1 or 2")
	ErrInvalidMode         = errors.New("invalid mode")
	ErrMissingName         = errors.New("missing name after slash")
	ErrInvalidNamePassword = errors.New("name/password is not allowed to contain slashes")
	ErrInvalidValue        = fmt.Errorf("invalid value")
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
	username string
}

// NewStreamID creates new StreamID
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
// The accepted old stream id format is <mode>/<password>/<password>. The second slash and password is
// optional and defaults to empty. The new format is `#!::m=(request|publish),r=(stream-key),u=(username),s=(password)`
// If error is not nil then StreamID will remain unchanged.
func (s *StreamID) FromString(src string) error {

	if strings.HasPrefix(src, IDPrefix) {
		for _, kv := range strings.Split(src[len(IDPrefix):], ",") {
			kv2 := strings.SplitN(kv, "=", 2)
			if len(kv2) != 2 {
				return ErrInvalidValue
			}

			key, value := kv2[0], kv2[1]

			switch key {
			case "u":
				s.username = value

			case "r":
				s.name = value

			case "h":

			case "s":
				s.password = value

			case "t":

			case "m":
				switch value {
				case "request":
					s.mode = ModePlay

				case "publish":
					s.mode = ModePublish

				default:
					return ErrInvalidMode
				}

			// Ignore keys sent by Blackmagic Atem Mini Pro
			case "bmd_uuid":

			case "bmd_name":

			default:
				return fmt.Errorf("unsupported key '%s'", key)
			}
		}
	} else {
		split := strings.Split(src, "/")

		s.password = ""
		if len(split) == 3 {
			s.password = split[2]
		} else if len(split) != 2 {
			return ErrInvalidSlashes
		}
		modeStr := split[0]
		s.name = split[1]

		switch modeStr {
		case "play":
			s.mode = ModePlay
		case "publish":
			s.mode = ModePublish
		default:
			return ErrInvalidMode
		}
	}

	if len(s.name) == 0 {
		return ErrMissingName
	}

	s.str = src
	return nil
}

// toString returns a string representation of the streamid
func (s *StreamID) toString() (string, error) {
	mode := ""
	switch s.mode {
	case ModePlay:
		mode = "play"
	case ModePublish:
		mode = "publish"
	default:
		return "", ErrInvalidMode
	}
	if strings.Contains(s.name, "/") {
		return "", ErrInvalidNamePassword
	}
	if strings.Contains(s.password, "/") {
		return "", ErrInvalidNamePassword
	}
	if len(s.password) == 0 {
		return fmt.Sprintf("%s/%s", mode, s.name), nil
	}
	return fmt.Sprintf("%s/%s/%s", mode, s.name, s.password), nil
}

// Match checks a streamid against a string with wildcards.
// The string may contain * to match any number of characters.
func (s StreamID) Match(pattern string) bool {
	return wildcard.Match(pattern, s.str)
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

func (s StreamID) Username() string {
	return s.username
}
