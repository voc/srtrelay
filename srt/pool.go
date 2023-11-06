package srt

import "sync"

func newBufferPool(packetSize uint) *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			buf := make([]byte, packetSize)
			return &buf
		},
	}
}
