package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	const numWorkers = 1024 * 1024
	const bufferSize = 1024
	var numCalcsCreated int
	var wg sync.WaitGroup

	calcPool := &sync.Pool{
		New: func() any {
			numCalcsCreated += 1
			mem := make([]byte, bufferSize)
			return &mem
		},
	}

	// Seed the pool with four buffers.
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())

	wg.Add(numWorkers)
	for i := numWorkers; i > 0; i-- {
		go func() {
			defer wg.Done()

			mem := calcPool.Get().(*[]byte)
			defer calcPool.Put(mem)

			// Assume something interesting, but quick is being done with this memory.
			// ...
			time.Sleep(1)
		}()
	}

	wg.Wait()
	fmt.Printf("%d calculators were created.\n", numCalcsCreated)
}
