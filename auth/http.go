package auth

import "github.com/voc/srtrelay/stream"

type httpAuth struct {
	url string
}

type HttpAuthConfig struct {
	URL string
}

func NewHttpAuth(config HttpAuthConfig) *httpAuth {
	return &httpAuth{
		url: config.URL,
	}
}

// Implement Authenticator
func (h *httpAuth) Authenticate(streamid stream.StreamID) bool {
	// TODO: implement post request
	return false
}
