package mpegts

import (
	"bytes"
	"io"
	"testing"
)

// Mock codec storing all parsed data
type mockCodec struct {
	data []byte
	done chan struct{}
	err  error
}

func (m *mockCodec) HasInit() bool {
	return false
}

func (m *mockCodec) InitPacket() ([]byte, error) {
	return []byte{}, nil
}

func (m *mockCodec) Parse(rd *io.PipeReader) {
	tmp := make([]byte, MaxPayloadSize)
	var n int
	n, m.err = rd.Read(tmp)
	m.data = append(m.data, tmp[:n]...)
	rd.Close()
	m.done <- struct{}{}
}

func (m *mockCodec) Data() ([]byte, error) {
	<-m.done
	return m.data, m.err
}

// Test PES encoding against our PES parser
func TestPES_encodeVideoPES(t *testing.T) {
	parser := &mockCodec{
		done: make(chan struct{}),
	}
	es := ElementaryStream{
		parser: parser,
	}

	expected := []byte{1, 2, 3, 4}
	payload, err := encodeVideoPES(expected)

	if err != nil {
		t.Fatal("Encode failed", err)
	}

	pkt := Packet{
		PID:     256,
		PUSI:    true,
		Payload: payload,
	}

	err = es.ParsePES(&pkt)
	if err != nil {
		t.Fatal("Parse failed", err)
	}

	res, err := parser.Data()
	if err != nil {
		t.Fatal("Mock Parser failed", err)
	}
	if bytes.Compare(res, expected) != 0 {
		t.Errorf("Got payload %v, expected %v", res, expected)
	}
}
