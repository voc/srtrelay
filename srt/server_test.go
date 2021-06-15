package srt

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/voc/srtrelay/relay"
)

func compareStats(got, expected []*relay.StreamStatistics) error {
	if len(got) != len(expected) {
		return fmt.Errorf("Wrong number of streams: got %d, expected %d", len(got), len(expected))
	}
	for i, v := range expected {
		if !reflect.DeepEqual(*v, *got[i]) {
			return fmt.Errorf("Invalid stream statistics: got %v, expected %v", *got[i], *v)
		}
	}
	return nil
}

func TestServerImpl_GetStatistics(t *testing.T) {
	r := relay.NewRelay(&relay.RelayConfig{})
	s := &ServerImpl{
		relay:  r,
		config: &ServerConfig{Addresses: []string{"127.0.0.1:1337", "[::1]:1337"}},
	}
	r.Publish("s1")
	r.Subscribe("s1")
	r.Subscribe("s1")
	streams := s.GetStatistics()

	expected := []*relay.StreamStatistics{
		{Name: "s1", URL: "srt://127.0.0.1:1337?streamid=play/s1", Clients: 2, Created: streams[0].Created},
	}
	if err := compareStats(streams, expected); err != nil {
		t.Error(err)
	}
}
