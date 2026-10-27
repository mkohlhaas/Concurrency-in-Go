package main

import (
	"fmt"
	"sync"
)

const (
	// numWorkers is the number of goroutines to spawn, each of which will
	// request a buffer from the pool.
	numWorkers = 1024 * 1024

	// bufferSize is the size of each byte buffer allocated when the pool needs
	// to create a new one.
	bufferSize = 1024
)

func main() {
	// numBuffersCreated tracks how many times the pool's New function was
	// called. This is used to demonstrate how many allocations were saved by
	// reusing buffers.
	var numBuffersCreated int

	// wg is used to wait for all worker goroutines to complete.
	var wg sync.WaitGroup

	// bufferPool maintains a cache of reusable byte buffers. When Get() is
	// called and the pool is empty, New() is invoked to allocate a buffer.
	bufferPool := &sync.Pool{
		New: func() any {
			// Increment the allocation counter each time a new buffer is created.
			numBuffersCreated++
			mem := make([]byte, bufferSize)
			// Return a pointer to the slice so the header can be reused.
			return &mem
		},
	}

	// Seed the pool with four buffers to avoid initial allocations when
	// the first goroutines begin requesting buffers.
	bufferPool.Put(bufferPool.New())
	bufferPool.Put(bufferPool.New())
	bufferPool.Put(bufferPool.New())
	bufferPool.Put(bufferPool.New())

	// Spawn numWorkers goroutines. Each goroutine retrieves a buffer from
	// the pool, uses it, and returns it before exiting.
	for range numWorkers {
		wg.Go(func() {
			// Get a buffer from the pool. If none are available, a new one
			// will be created via the New function.
			mem := bufferPool.Get().(*[]byte)
			// Ensure the buffer is returned to the pool when this goroutine exits.
			defer bufferPool.Put(mem)

			// Assume something interesting, but quick is being done with this memory.
			// ...
		})
	}

	// Block until all goroutines have completed.
	wg.Wait()

	// Report how many buffers were actually allocated versus how many were reused.
	fmt.Printf("%d buffers were created.\n", numBuffersCreated)
}
