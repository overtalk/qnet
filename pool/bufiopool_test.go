package pool_test

import (
	"bytes"
	"testing"

	"github.com/overtalk/qnet/pool"
)

func TestBufReader(t *testing.T) {
	readerPool := pool.NewBufReaderPool(100, 1<<10)
	bytesReader := &bytes.Buffer{}
	rdr := readerPool.Get(bytesReader)
	rdr.Free()
}

func BenchmarkBufReader(b *testing.B) {
	readerPool := pool.NewBufReaderPool(100, 1<<10)
	for i := 0; i < b.N; i++ {
		bytesReader := &bytes.Buffer{}
		rdr := readerPool.Get(bytesReader)
		rdr.Free()
	}
}
