package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

type value struct {
	mu    sync.Mutex
	value int
}

func main() {
	printSum := func(v1, v2 *value) {
		defer wg.Done()
		v1.mu.Lock()
		defer v1.mu.Unlock()

		// This makes the deadlock possible, but doesnt't guarantee it.
		time.Sleep(2 * time.Second) // simulation of long-running process

		v2.mu.Lock()
		defer v2.mu.Unlock()
		fmt.Printf("sum=%v\n", v1.value+v2.value)
	}

	var a, b value
	wg.Add(2)

	// The first call to printSum locks a and then attempts to lock b, but in the
	// meantime our second call to printSum has locked b and has attempted to lock
	// a. Both goroutines wait infinitely on each other. Deadlock!
	go printSum(&a, &b)
	go printSum(&b, &a)

	wg.Wait()
}
