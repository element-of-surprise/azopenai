package rest

import (
	"bytes"
	"sync"
)

type bufferPool struct {
	buffers chan *bytes.Buffer
	pool    *sync.Pool
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		buffers: make(chan *bytes.Buffer, 100),
		pool: &sync.Pool{
			New: func() any {
				return &bytes.Buffer{}
			},
		},
	}
}

func (b bufferPool) Get() *bytes.Buffer {
	select {
	case buff := <-b.buffers:
		return buff
	default:
	}
	return b.pool.Get().(*bytes.Buffer)
}

func (b bufferPool) Put(buff *bytes.Buffer) {
	buff.Reset()
	select {
	case b.buffers <- buff:
		return
	default:
	}
	b.pool.Put(buff)
}
