package main

import (
	"fmt"
	"sync"
)

func main() {
	// Create a new sync.Pool. The New function is called lazily when
	// Get() is called and the pool has no available objects.
	myPool := &sync.Pool{
		New: func() any {
			fmt.Println("Creating new instance.")
			return struct{}{}
		},
	}

	// Get() from an empty pool calls New() to create a new instance.
	// This instance is not Put() back, so it is lost to the pool.
	myPool.Get()

	// Pool is empty again, so New() is called again to create another instance.
	// We store this instance in a variable so we can return it to the pool.
	instance := myPool.Get()

	// Put() returns the instance to the pool, making it available for reuse.
	myPool.Put(instance)

	// Get() retrieves the instance we just Put() back. New() is not called.
	// This instance is not Put() back, so it is lost for good.
	myPool.Get()
}
