package pool_test

import (
	"testing"

	zeropool "github.com/overtalk/qnet/pool"
)

func TestByteBufPool(t *testing.T) {
	pool := zeropool.NewByteBufPool(100)
	b1 := pool.Get()
	b2 := pool.Get()
	b1.WriteString("hello")
	b2.WriteString("world")
	t.Logf("b1: %v\n", b1.String())
	t.Logf("b2: %v\n", b2.String())
	b1.Free()
	b2.Free()

	b3 := pool.Get()
	b4 := pool.Get()
	t.Logf("b3: %v\n", b3.String())
	t.Logf("b4: %v\n", b4.String())
	b3.Free()
	b4.Free()
}

func TestZlibPool(t *testing.T) {
	pool := zeropool.NewZlibWriterPool(100)
	b1 := pool.Get()
	b2 := pool.Get()
	r1 := b1.Compress([]byte("hello, world!"))
	r2 := b2.Compress([]byte("how old are you?"))
	t.Logf("b1: %v\n", r1)
	t.Logf("b2: %v\n", r2)
	b1.Free()
	b2.Free()

	b3 := pool.Get()
	b4 := pool.Get()
	r3 := b3.Compress([]byte("hello, world!"))
	r4 := b4.Compress([]byte("how old are you?"))
	t.Logf("b3: %v\n", r3)
	t.Logf("b4: %v\n", r4)
	b3.Free()
	b4.Free()
}
