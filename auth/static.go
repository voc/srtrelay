package auth

import (
	"github.com/voc/srtrelay/stream"
)

type StaticAuth struct {
	allow []string
}

type StaticAuthConfig struct {
	Allow []string
}

// NewStaticAuth creates an Authenticator with a static config backend
func NewStaticAuth(config StaticAuthConfig) *StaticAuth {
	return &StaticAuth{
		allow: config.Allow,
	}
}

// Implement Authenticator

// Authenticate tries to match the stream id against the locally
// configured matches in the allowlist.
func (auth *StaticAuth) Authenticate(streamid stream.StreamID) bool {
	for _, allowed := range auth.allow {
		if streamid.Match(allowed) {
			return true
		}
	}
	return false
}
