package pebble

import (
	"bytes"
	"sync"
)

type ByteBufferPool struct {
	pool sync.Pool
}

func NewByteBufferPool() *ByteBufferPool {
	bbp := &ByteBufferPool{}
	bbp.pool.New = func() any {
		return bytes.NewBuffer(nil)
	}
	return bbp
}

func (bbp *ByteBufferPool) Get() *bytes.Buffer {
	return bbp.pool.Get().(*bytes.Buffer)
}

func (bbp *ByteBufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	bbp.pool.Put(buf)
}

type ByteArrayPool struct {
	pool sync.Pool
}

func NewArrayBufferPool() *ByteArrayPool {
	bbp := &ByteArrayPool{}
	bbp.pool.New = func() any {
		return make([]byte, 0)
	}
	return bbp
}

func (bbp *ByteArrayPool) Get() []byte {
	return bbp.pool.Get().([]byte)
}

func (bbp *ByteArrayPool) Put(buf []byte) {
	buf = buf[:0]
	bbp.pool.Put(buf)
}
