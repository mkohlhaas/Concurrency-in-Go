package main

import (
	"fmt"
	"sync"
)

const numIncs = 100

func main() {
	var count int

	increment := func() {
		count++
	}

	var once sync.Once

	var wg sync.WaitGroup
	wg.Add(numIncs)

	for range numIncs {
		go func() {
			defer wg.Done()
			once.Do(increment)
		}()
	}

	wg.Wait()
	fmt.Printf("Count: %d\n", count)
}
