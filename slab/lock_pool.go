package slab

import (
	"reflect"
	"sync"
	"unsafe"
)

// LockPool is a lock-free slab allocation memory pool.
type LockPool struct {
	pages   []lockPage
	minSize int
	maxSize int
}

// NewLockPool create a lock-free slab allocation memory pool.
// minSize is the smallest chunk size.
// maxSize is the lagest chunk size.
// factor is used to control growth of chunk size.
// pageSize is the memory size of each slab class.
func NewLockPool(minSize, maxSize, factor, pageSize int) *LockPool {
	n := 0
	for chunkSize := minSize; chunkSize <= maxSize && chunkSize <= pageSize; chunkSize *= factor {
		n++
	}
	pool := &LockPool{make([]lockPage, n), minSize, maxSize}
	n = 0
	for chunkSize := minSize; chunkSize <= maxSize && chunkSize <= pageSize; chunkSize *= factor {
		c := &pool.pages[n]
		c.size = chunkSize
		c.page = make([]byte, pageSize)
		c.chunks = make([][]byte, pageSize/chunkSize)
		c.head = 0
		c.tail = pageSize/chunkSize - 1
		for i := 0; i < len(c.chunks); i++ {
			// lock down the capacity to protect append operation
			c.chunks[i] = c.page[i*chunkSize : (i+1)*chunkSize : (i+1)*chunkSize]
			if i == len(c.chunks)-1 {
				c.pageBegin = uintptr(unsafe.Pointer(&c.page[0]))
				c.pageEnd = uintptr(unsafe.Pointer(&c.chunks[i][0]))
			}
		}
		n++
		// update maxSize
		pool.maxSize = chunkSize
	}
	return pool
}

// Alloc try alloc a []byte from internal slab class if no free chunk in slab class Alloc will make one.
func (pool *LockPool) Alloc(size int) []byte {
	if size <= pool.maxSize {
		for i := 0; i < len(pool.pages); i++ {
			if pool.pages[i].size >= size {
				mem := pool.pages[i].Pop()
				if mem != nil {
					return mem[:size]
				}
				break
			}
		}
	}
	return make([]byte, size)
}

// Free release a []byte that alloc from Pool.Alloc.
func (pool *LockPool) Free(mem []byte) {
	if size := cap(mem); size <= pool.maxSize {
		for i := 0; i < len(pool.pages); i++ {
			if pool.pages[i].size == size {
				pool.pages[i].Push(mem)
				break
			}
		}
	}
}

type lockPage struct {
	sync.Mutex
	size      int
	page      []byte
	pageBegin uintptr
	pageEnd   uintptr
	chunks    [][]byte
	head      int
	tail      int
}

func (c *lockPage) Push(mem []byte) {
	ptr := (*reflect.SliceHeader)(unsafe.Pointer(&mem)).Data
	if c.pageBegin <= ptr && ptr <= c.pageEnd {
		c.Lock()
		c.tail++
		n := c.tail % len(c.chunks)
		if c.chunks[n] != nil {
			panic("slab.LockPool: Double Free")
		}
		c.chunks[n] = mem
		c.Unlock()
	}
}

func (c *lockPage) Pop() []byte {
	var mem []byte
	c.Lock()
	if c.head <= c.tail {
		n := c.head % len(c.chunks)
		mem = c.chunks[n]
		c.chunks[n] = nil
		c.head++
	}
	c.Unlock()
	return mem
}
