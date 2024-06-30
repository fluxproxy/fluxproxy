package internal

import (
	"sync"
)

type ByteBufferPool struct {
	bufferSize int
	pool       *sync.Pool
}

func NewByteBufferPool(bufferSize int) *ByteBufferPool {
	return &ByteBufferPool{
		bufferSize,
		&sync.Pool{
			New: func() interface{} { return make([]byte, 0, bufferSize) }},
	}
}

func (sf *ByteBufferPool) Get() []byte {
	return sf.pool.Get().([]byte)
}

func (sf *ByteBufferPool) Put(b []byte) {
	if cap(b) != sf.bufferSize {
		panic("illegal buffer: invalid buffer size")
	}
	sf.pool.Put(b[:0]) //nolint: staticcheck
}
