package auth

import (
	"testing"

	"github.com/voc/srtrelay/stream"
)

func TestStaticAuth_Authenticate(t *testing.T) {
	tests := []struct {
		name     string
		allow    []string
		streamid string
		want     bool
	}{
		{"MatchFirst", []string{"play/*", ""}, "play/foobar", true},
		{"MatchLast", []string{"", "play/*"}, "play/foobar", true},
		{"MatchNone", []string{}, "play/foobar", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewStaticAuth(StaticAuthConfig{
				Allow: tt.allow,
			})
			streamid := stream.StreamID{}
			if err := streamid.FromString(tt.streamid); err != nil {
				t.Error(err)
			}
			if got := auth.Authenticate(streamid); got != tt.want {
				t.Errorf("StaticAuth.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}
