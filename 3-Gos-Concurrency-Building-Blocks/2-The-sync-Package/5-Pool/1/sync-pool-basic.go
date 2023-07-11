package main

import (
	"fmt"
	"sync"
)

func main() {
	myPool := &sync.Pool{
		New: func() any {
			fmt.Println("Creating new instance.")
			return struct{}{}
		},
	}

	myPool.Get()             // will never be put back in the pool
	instance := myPool.Get() // creates another instance
	myPool.Put(instance)     // and puts it back
	myPool.Get()             // will also be lost for good
}
