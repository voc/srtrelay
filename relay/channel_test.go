package relay

import (
	"reflect"
	"testing"
)

func TestChannel_PubSub(t *testing.T) {
	ch := NewChannel()

	// sub
	out, unsub := ch.Sub()
	data := []byte{1, 2, 3, 4}

	// pub
	ch.Pub(data)
	got := <-out

	if !reflect.DeepEqual(got, data) {
		t.Errorf("Sub ret = %x, want %x", got, data)
	}

	// unsub
	unsub()

	// pub2
	ch.Pub(data)
	got = <-out

	if got != nil {
		t.Errorf("Read after unsub ret %x, want nil", got)
	}
}

func TestChannel_DropOnOverflow(t *testing.T) {
	ch := NewChannel()

	sub, _ := ch.Sub()
	capacity := cap(sub) + 1

	// Overflow subscriber on purpose
	for i := 0; i < capacity; i++ {
		ch.Pub([]byte{})
		// Should drop and close subscriber on last iteration
	}

	// Check removal
	if remaining := len(ch.subs); remaining > 0 {
		t.Errorf("Got %d remaining channels, expected 0", remaining)
	}

	// Read elements from buffer
	for i := 0; i < capacity; i++ {
		expected := (i < capacity-1)
		_, got := <-sub
		if got != expected {
			t.Errorf("Channel read expects %t on pos %d, got %t", expected, i, got)
		}
	}
}

func TestChannel_Close(t *testing.T) {
	ch := NewChannel()
	sub1, _ := ch.Sub()
	_, _ = ch.Sub()

	if got := len(ch.subs); got != 2 {
		t.Fatalf("Expected 2 subscribers on channel, got %d", got)
	}

	ch.Close()

	if got := len(ch.subs); got != 0 {
		t.Errorf("Expected 0 subscribers after close, got %d", got)
	}

	if _, ok := <-sub1; ok {
		t.Error("Subscriber channel should be closed after Close")
	}
}
