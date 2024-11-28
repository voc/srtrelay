package srt

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/haivision/srtgo"
	"github.com/voc/srtrelay/relay"
	"github.com/voc/srtrelay/stream"
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
	r := relay.NewRelay(&relay.RelayConfig{
		BufferSize: 1,
		PacketSize: 1,
	})
	s := &ServerImpl{
		relay:  r,
		config: &ServerConfig{Addresses: []string{"127.0.0.1:1337", "[::1]:1337"}, PublicAddress: "testserver.de:1337"},
	}
	if _, err := r.Publish("s1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.Subscribe("s1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.Subscribe("s1"); err != nil {
		t.Fatal(err)
	}
	streams := s.GetStatistics()

	expected := []*relay.StreamStatistics{
		//{Name: "s1", URL: "srt://testserver.de:1337?streamid=play/s1", Clients: 2, Created: streams[0].Created},
		{Name: "s1", URL: "srt://testserver.de:1337?streamid=#!::m=request,r=s1", Clients: 2, Created: streams[0].Created},
	}
	if err := compareStats(streams, expected); err != nil {
		t.Error(err)
	}
}

type testSocket struct {
	N  int
	ch chan []byte

	numWritten int
}

func (s *testSocket) Read(b []byte) (int, error) {
	buf, ok := <-s.ch
	if !ok {
		return 0, io.EOF
	}
	length := copy(b, buf)
	return length, nil
}

func (s *testSocket) Write(b []byte) (int, error) {
	s.numWritten++
	return len(b), nil
}

func (s *testSocket) Close() {}

func (s *testSocket) Stats() (*srtgo.SrtStats, error) {
	return &srtgo.SrtStats{}, nil
}

func TestPublish(t *testing.T) {
	s := NewServer(&Config{
		Server: ServerConfig{},
		Relay:  relay.RelayConfig{BufferSize: 50, PacketSize: 1316},
	})

	rd := testSocket{ch: make(chan []byte)}
	wr := testSocket{}
	id, err := stream.NewStreamID("test", "", stream.ModePublish)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// pub routine
	go func() {
		defer wg.Done()
		err = s.publish(&srtConn{
			socket:   &rd,
			streamid: id,
			address:  "publisher:1234",
		})
		if err != io.EOF {
			t.Error("publisher error", err)
		}
	}()

	// sub routine
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		err := s.play(&srtConn{
			socket:   &wr,
			streamid: id,
			address:  "player:1234",
		})
		if err != nil {
			t.Error("player error", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 100; i++ {
		rd.ch <- []byte{1, 2, 3, 4}
		time.Sleep(1 * time.Millisecond)
	}
	close(rd.ch)
	wg.Wait()
	if wr.numWritten != 100 {
		t.Errorf("Wrong number of packets written: got %d, expected %d", wr.numWritten, 100)
	}
}
