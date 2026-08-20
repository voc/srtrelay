package relay

import "testing"

func TestBuffer_SendThenRetainPanics(t *testing.T) {
	pool := NewBufferPool(1)
	buf := pool.Get()

	defer func() {
		if recover() == nil {
			t.Fatal("expected send-then-retain ordering to panic after subscriber release")
		}
	}()

	buf.Release()
	buf.Retain()
}
