package pool

import (
	"bytes"
	"compress/zlib"
)

// ByteBuf a bytes buffer
type ByteBuf struct {
	*bytes.Buffer
	pool *ByteBufPool
}

// NewByteBuf create a ByteBuf struct
func NewByteBuf(buf *bytes.Buffer) *ByteBuf {
	return &ByteBuf{Buffer: buf, pool: nil}
}

// Free free this ByteBuf
func (buf *ByteBuf) Free() {
	if buf.pool != nil {
		buf.pool.put(buf)
	}
}

// ByteBufPool a *ByteBuf pool
type ByteBufPool struct {
	pool chan *ByteBuf
}

// NewByteBufPool creates a *ByteBuf pool
func NewByteBufPool(size int) *ByteBufPool {
	return &ByteBufPool{pool: make(chan *ByteBuf, size)}
}

// Get returns a bytes.Buff or creata a new one if not enough
func (bp *ByteBufPool) Get() *ByteBuf {
	var buf *ByteBuf
	select {
	case buf = <-bp.pool:
		buf.Reset()
	default:
		buf = NewByteBuf(&bytes.Buffer{})
		buf.pool = bp
	}
	return buf
}

// Put puts back it to the pool
func (bp *ByteBufPool) put(buf *ByteBuf) {
	if buf != nil {
		select {
		case bp.pool <- buf:
		default:
			// do nothing, just discard
		}
	}
}

// ZlibWriter a zlib writer containg a writer an a buffer
type ZlibWriter struct {
	buffer *bytes.Buffer
	writer *zlib.Writer

	pool *ZlibWriterPool
}

// NewZlibWriter create a zlibWriter struct
func NewZlibWriter() *ZlibWriter {
	buffer := new(bytes.Buffer)
	writer := zlib.NewWriter(buffer)
	return &ZlibWriter{buffer, writer, nil}
}

// Reset reset the underlying buffer and writer
func (w *ZlibWriter) Reset() {
	w.buffer.Reset()
	w.writer.Reset(w.buffer)
}

// Compress compress a raw bytes of data
func (w *ZlibWriter) Compress(b []byte) []byte {
	w.writer.Write(b)
	w.writer.Close()
	return w.buffer.Bytes()
}

// Free free its underlying resource
func (w *ZlibWriter) Free() {
	if w.pool != nil {
		w.pool.put(w)
	}
}

// ZlibWriterPool a *ZlibWriter pool
type ZlibWriterPool struct {
	pool chan *ZlibWriter
}

// NewZlibWriterPool create a ZlibWriterPool struct
func NewZlibWriterPool(size int) *ZlibWriterPool {
	return &ZlibWriterPool{pool: make(chan *ZlibWriter, size)}
}

// Get get a free *ZlibIem from the pool
func (zp *ZlibWriterPool) Get() *ZlibWriter {
	var w *ZlibWriter
	select {
	case w = <-zp.pool:
		w.Reset()
	default:
		w = NewZlibWriter()
		w.pool = zp
	}
	return w
}

// Put put a free *ZlibWriter to the pool and if full pool, discard it.
func (zp *ZlibWriterPool) put(w *ZlibWriter) {
	if w != nil {
		select {
		case zp.pool <- w:
		default:
			// do nothing, just discard
		}
	}
}
