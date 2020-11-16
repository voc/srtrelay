package auth

import (
	"github.com/voc/srtrelay/stream"
)

type staticAuth struct {
	allow []string
}

type StaticAuthConfig struct {
	Allow []string
}

func NewStaticAuth(config StaticAuthConfig) *staticAuth {
	return &staticAuth{
		allow: config.Allow,
	}
}

// Implement Authenticator
func (auth *staticAuth) Authenticate(streamid stream.StreamID) bool {
	for _, allowed := range auth.allow {
		if streamid.Match(allowed) {
			return true
		}
	}
	return false
}
