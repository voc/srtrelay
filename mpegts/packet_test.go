package mpegts

import (
	"encoding/hex"
	"testing"
)

func TestPacket_ToBytes_FromBytes(t *testing.T) {
	buf := make([]byte, PacketLen)
	payload := []byte{1, 2, 3, 4, 6}

	// pad with adaptationField
	adaptationLen := MaxPayloadSize - len(payload) - 1
	adaptationField := make([]byte, adaptationLen)
	adaptationField[0] = 0x3 << 6
	for i := 1; i < adaptationLen; i++ {
		adaptationField[i] = 0xff
	}

	// encode packet
	pkt1 := Packet{
		PID:             0x100,
		Payload:         payload,
		AdaptationField: adaptationField,
		PUSI:            true,
	}
	err := pkt1.ToBytes(buf)
	if err != nil {
		t.Fatal(err)
	}

	// parse packet
	pkt2 := Packet{}
	err = pkt2.FromBytes(buf)
	if err != nil {
		t.Fatal(err, hex.Dump(buf))
	}

	if pkt1.PID != pkt2.PID {
		t.Errorf("Failed to encode/parse PID, got: %d, expected: %d", pkt2.PID, pkt1.PID)
	}

	if pkt1.PUSI != pkt2.PUSI {
		t.Errorf("Failed to encode/parse PUSI, got: %v, expected: %v", pkt2.PUSI, pkt1.PUSI)
	}

	if hex.EncodeToString(pkt1.Payload) != hex.EncodeToString(pkt2.Payload) {
		t.Errorf("Failed to encode/parse Payload,\n got: %s,\n expected %s",
			hex.EncodeToString(pkt2.Payload),
			hex.EncodeToString(pkt1.Payload))
	}

	if hex.EncodeToString(pkt1.AdaptationField) != hex.EncodeToString(pkt2.AdaptationField) {
		t.Errorf("Failed to encode/parse AdaptationField,\n got: %s,\n expected %s",
			hex.EncodeToString(pkt2.AdaptationField),
			hex.EncodeToString(pkt1.AdaptationField))
	}
}
