package auth

import "github.com/voc/srtrelay/stream"

type Authenticator interface {
	Authenticate(stream.StreamID) bool
}
