package main

import (
	"fmt"
	"sync"
)

const runs = 1_000

func main() {
	var count int
	var lock sync.Mutex

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

	var arithmetic sync.WaitGroup

	for range runs {
		arithmetic.Go(increment)
	}

	for range runs {
		arithmetic.Go(decrement)
	}

	arithmetic.Wait()

	fmt.Println("Arithmetic complete.")
	fmt.Printf("Count: %d\n", count)
}
