package rest

import (
	"io"
	"sync"
)

// buffer is a buffer object that implements the io.ReadCloser interface.
// It is more efficient than bytes.buffer for our purposes, as we don't need
// to write to the buffer after it has been created.
type buffer struct {
	b   []byte
	ptr int
}

// Read implements the io.Reader interface.
func (b *buffer) Read(p []byte) (n int, err error) {
	if b.ptr >= len(b.b) {
		return 0, io.EOF
	}
	n = copy(p, b.b[b.ptr:])
	b.ptr += n
	return
}

// Close implements the io.Closer interface.
func (b *buffer) Close() error {
	b.b = nil
	b.ptr = 0
	return nil
}

// Reset resets the buffer to the given byte slice.
func (b *buffer) Reset(buff []byte) {
	b.b = buff
	b.ptr = 0
}

type bufferPool struct {
	buffers chan *buffer
	pool    *sync.Pool
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		buffers: make(chan *buffer, 100),
		pool: &sync.Pool{
			New: func() any {
				return &buffer{}
			},
		},
	}
}

func (b bufferPool) Get() *buffer {
	select {
	case buff := <-b.buffers:
		return buff
	default:
	}
	return b.pool.Get().(*buffer)
}

func (b bufferPool) Put(buff *buffer) {
	select {
	case b.buffers <- buff:
		return
	default:
	}
	b.pool.Put(buff)
}
