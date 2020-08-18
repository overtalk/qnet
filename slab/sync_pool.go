package slab

import "sync"

// SyncPool is a sync.Pool base slab allocation memory pool
type SyncPool struct {
	pages     []sync.Pool
	pagesSize []int
	minSize   int
	maxSize   int
}

// NewSyncPool create a sync.Pool base slab allocation memory pool.
// minSize is the smallest chunk size.
// maxSize is the lagest chunk size.
// factor is used to control growth of chunk size.
func NewSyncPool(minSize, maxSize, factor int) *SyncPool {
	n := 0
	for chunkSize := minSize; chunkSize <= maxSize; chunkSize *= factor {
		n++
	}
	pool := &SyncPool{
		make([]sync.Pool, n),
		make([]int, n),
		minSize, maxSize,
	}
	n = 0
	for chunkSize := minSize; chunkSize <= maxSize; chunkSize *= factor {
		pool.pagesSize[n] = chunkSize
		pool.pages[n].New = func(size int) func() interface{} {
			return func() interface{} {
				buf := make([]byte, size)
				return &buf
			}
		}(chunkSize)
		n++
		// update maxSize
		pool.maxSize = chunkSize
	}
	return pool
}

// Alloc try alloc a []byte from internal slab class if no free chunk in slab class Alloc will make one.
func (pool *SyncPool) Alloc(size int) []byte {
	if size <= pool.maxSize {
		for i := 0; i < len(pool.pagesSize); i++ {
			if pool.pagesSize[i] >= size {
				mem := pool.pages[i].Get().(*[]byte)
				return (*mem)[:size]
			}
		}
	}
	return make([]byte, size)
}

// Free release a []byte that alloc from Pool.Alloc.
func (pool *SyncPool) Free(mem []byte) {
	if size := cap(mem); size <= pool.maxSize {
		for i := 0; i < len(pool.pagesSize); i++ {
			if pool.pagesSize[i] >= size {
				pool.pages[i].Put(&mem)
				return
			}
		}
	}
}
