package main

import (
	"fmt"
	"sync"
)

const numGreeters = 50

func main() {
	var wg sync.WaitGroup

	hello := func(id int) {
		defer wg.Done()
		fmt.Printf("Hello from %v!\n", id)
	}

	wg.Add(numGreeters)

	for i := range numGreeters {
		go hello(i + 1)
	}

	wg.Wait()
}
