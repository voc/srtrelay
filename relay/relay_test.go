package relay

import (
	"reflect"
	"testing"
	"time"
)

func TestRelayImpl_SubscribeAndUnsubscribe(t *testing.T) {
	config := RelayConfig{BufferSize: 50, PacketSize: 1}
	relay := NewRelay(&config)
	data := []byte{1, 2, 3, 4}

	pub, err := relay.Publish("test")
	if err != nil {
		t.Fatal(err)
	}

	sub, unsub, err := relay.Subscribe("test")
	if err != nil {
		t.Fatal(err)
	}

	// send
	pub <- data

	// receive
	got, ok := <-sub
	if !ok {
		t.Fatal("Subscriber channel should not be closed")
	}
	if !reflect.DeepEqual(got, data) {
		t.Errorf("Sub ret = %x, want %x", got, data)
	}

	// unsubscribe
	unsub()

	// 2nd send
	pub <- data
	got, ok = <-sub

	if got != nil || ok {
		t.Errorf("Read after unsub ret %x, want nil", got)
	}
}

func TestRelayImpl_PublisherClose(t *testing.T) {
	config := RelayConfig{BufferSize: 1, PacketSize: 1}
	relay := NewRelay(&config)

	ch, _ := relay.Publish("test")
	sub, unsub, _ := relay.Subscribe("test")
	close(ch)

	// Wait for async teardown in goroutine
	time.Sleep(100 * time.Millisecond)

	if _, ok := <-sub; ok {
		t.Error("Subscriber channel should be closed")
	}

	// unsub after close shouldn't break
	unsub()

	_, err := relay.Publish("test")
	if err != nil {
		t.Error("Publish should be possible again after close")
	}
}

func TestRelayImpl_DoublePublish(t *testing.T) {
	config := RelayConfig{BufferSize: 1, PacketSize: 1}
	relay := NewRelay(&config)
	relay.Publish("foo")
	_, err := relay.Publish("foo")

	if err != ErrStreamAlreadyExists {
		t.Errorf("Publish to existing stream should return '%s', got '%s'", ErrStreamAlreadyExists, err)
	}
}

func TestRelayImpl_SubscribeNonExisting(t *testing.T) {
	config := RelayConfig{BufferSize: 1, PacketSize: 1}
	relay := NewRelay(&config)

	_, _, err := relay.Subscribe("foobar")
	if err != ErrStreamNotExisting {
		t.Errorf("Subscribe to non-existing stream should return '%s', got '%s'", ErrStreamNotExisting, err)
	}
}

func TestRelayImpl_ChannelExists(t *testing.T) {
	config := RelayConfig{BufferSize: 1, PacketSize: 1}
	relay := NewRelay(&config)

	ok := relay.ChannelExists("test")
	if ok {
		t.Fatal("Channel should not exist before publishing")
	}

	_, err := relay.Publish("test")
	if err != nil {
		t.Fatal(err)
	}
	ok = relay.ChannelExists("test")
	if !ok {
		t.Fatal("Channel should exist after publishing")
	}
}
