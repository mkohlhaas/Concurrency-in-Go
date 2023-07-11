package main

import (
	"fmt"
	"sync"
)

func main() {
	const numGreeters = 50
	var wg sync.WaitGroup

	hello := func(id int) {
		defer wg.Done()
		fmt.Printf("Hello from %v!\n", id)
	}

	wg.Add(numGreeters)
	for i := 0; i < numGreeters; i++ {
		go hello(i + 1)
	}
	wg.Wait()
}
