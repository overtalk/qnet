package pool

import (
	"bufio"
	"io"
)

// A BufReader wrapper a bufio.Reader
type BufReader struct {
	*bufio.Reader
	pool *BufReaderPool
}

// Free free it to a pool
func (br *BufReader) Free() {
	// decrease the underlying netconn object holding
	br.Reset(nil)
	if br.pool != nil {
		br.pool.put(br)
	}
}

// BufReaderPool a *bufio.Reader pool
// bufio.Reader can decrease the io call.
type BufReaderPool struct {
	pool   chan *BufReader
	rdSize int
}

// NewBufReaderPool creates a BufReader pool
func NewBufReaderPool(poolSize, readSize int) *BufReaderPool {
	return &BufReaderPool{make(chan *BufReader, poolSize), readSize}
}

// Get returns a bytes.Buff or creata a new one if not enough
func (bp *BufReaderPool) Get(r io.Reader) *BufReader {
	var item *BufReader
	select {
	case item = <-bp.pool:
		item.Reset(r)
	default:
		item = &BufReader{
			Reader: bufio.NewReaderSize(r, bp.rdSize),
			pool:   bp,
		}
	}
	return item
}

// Put puts back it to the pool
func (bp *BufReaderPool) put(r *BufReader) {
	if r != nil {
		select {
		case bp.pool <- r:
		default:
			// do nothing, just discard
		}
	}
}
