package config

import (
	"net/netip"
	"slices"
	"testing"
	"time"

	"github.com/voc/srtrelay/auth"
	"gotest.tools/v3/assert"
)

func TestConfig(t *testing.T) {
	conf, err := Parse([]string{"testfiles/config_test.toml"})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, conf.App.Addresses[0], "127.0.0.1:5432")
	assert.Equal(t, conf.App.LatencyMs, uint(1337))
	assert.Equal(t, conf.App.Buffersize, uint(123000))
	assert.Equal(t, conf.App.SyncClients, true)
	assert.Equal(t, conf.App.PacketSize, uint(1456))
	assert.Equal(t, conf.App.LossMaxTTL, uint(50))
	assert.Equal(t, conf.App.PublicAddress, "dontlookmeup:5432")

	assert.Equal(t, conf.API.Enabled, false)
	assert.Equal(t, conf.API.Address, ":1234")

	assert.Equal(t, conf.Auth.Type, "http")
	assert.Equal(t, conf.Auth.Static.Allow[0], "play/*")
	assert.Equal(t, conf.Auth.HTTP.URL, "http://localhost:1235/publish")
	assert.Equal(t, conf.Auth.HTTP.Timeout, auth.Duration(time.Second*5))
	assert.Equal(t, conf.Auth.HTTP.Application, "foo")
	assert.Equal(t, conf.Auth.HTTP.PasswordParam, "pass")
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		expected    []netip.AddrPort
		expectedErr bool
	}{
		{"localhost", "localhost:1337", []netip.AddrPort{netip.MustParseAddrPort("127.0.0.1:1337"), netip.MustParseAddrPort("[::1]:1337")}, false},
		{"no host", ":1337", []netip.AddrPort{netip.MustParseAddrPort("0.0.0.0:1337"), netip.MustParseAddrPort("[::]:1337")}, false},
		{"all v4", "0.0.0.0:1337", []netip.AddrPort{netip.MustParseAddrPort("0.0.0.0:1337")}, false},
		{"all v6", "[::]:1337", []netip.AddrPort{netip.MustParseAddrPort("[::]:1337")}, false},
		{"v6", "[1234::beef]:1337", []netip.AddrPort{netip.MustParseAddrPort("[1234::beef]:1337")}, false},
		{"invalid", "localhost:abc", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddress(tt.addr)
			assert.Equal(t, err != nil, tt.expectedErr)
			ok := slices.EqualFunc(got, tt.expected, func(a, b netip.AddrPort) bool {
				return a.String() == b.String()
			})
			if !ok {
				t.Fatalf("got = %v, want %v", got, tt.expected)
			}
		})
	}
}
