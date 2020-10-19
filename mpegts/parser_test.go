package mpegts

import (
	"io/ioutil"
	"testing"
)

func TestParser_ParseH264(t *testing.T) {
	data, err := ioutil.ReadFile("h264.ts")
	if err != nil {
		t.Fatal("failed to open test file")
	}
	p := NewParser()
	err = p.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.hasInit() {
		t.Error("Should find init")
	}
	pkts, err := p.InitData()
	if err != nil {
		t.Fatalf("Init data failed: %v", err)
	}
	if pkts == nil {
		t.Error("Init should not be nil")
	}
	numPackets := len(pkts)
	if numPackets != 20 {
		t.Errorf("Expected 3 init packets, got %d", numPackets)
	}
}
