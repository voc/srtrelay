package config

import (
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
	assert.Equal(t, conf.App.Latency, uint(1337))
	assert.Equal(t, conf.App.Buffersize, uint(123000))
	assert.Equal(t, conf.App.SyncClients, true)
	assert.Equal(t, conf.App.PacketSize, uint(1456))
	assert.Equal(t, conf.App.LossMaxTTL, uint(50))
	assert.Equal(t, conf.App.ListenTimeout, uint(5555))
	assert.Equal(t, conf.App.PublicAddress, "dontlookmeup:5432")
	assert.Equal(t, conf.App.ListenBacklog, 30)

	assert.Equal(t, conf.API.Enabled, false)
	assert.Equal(t, conf.API.Address, ":1234")

	assert.Equal(t, conf.Auth.Type, "http")
	assert.Equal(t, conf.Auth.Static.Allow[0], "play/*")
	assert.Equal(t, conf.Auth.HTTP.URL, "http://localhost:1235/publish")
	assert.Equal(t, conf.Auth.HTTP.Timeout, auth.Duration(time.Second*5))
	assert.Equal(t, conf.Auth.HTTP.Application, "foo")
	assert.Equal(t, conf.Auth.HTTP.PasswordParam, "pass")
}
