package main

import (
	"fmt"
	"sync"
)

func main() {
	var count int
	var lock sync.Mutex
	var arithmetic sync.WaitGroup
	const runs = 1_000

	increment := func() {
		lock.Lock()
		defer lock.Unlock()
		count++
		fmt.Printf("Incrementing: %.3d\n", count)
	}

	decrement := func() {
		lock.Lock()
		defer lock.Unlock()
		count--
		fmt.Printf("Decrementing: %.3d\n", count)
	}

	// increment
	for i := 0; i <= runs; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			increment()
		}()
	}

	// decrement
	for i := 0; i <= runs; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			decrement()
		}()
	}

	arithmetic.Wait()
	fmt.Println("Arithmetic complete.")
	fmt.Printf("count: %d\n", count)
}
