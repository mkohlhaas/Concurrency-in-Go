package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var mu sync.Mutex
	cond := sync.NewCond(&mu) // 1. Initialize with a locker
	ready := false

	// Worker Goroutine
	go func() {
		mu.Lock()
		defer mu.Unlock()

		// 2. ALWAYS check the condition inside a 'for' loop
		for !ready {
			cond.Wait() // Suspends, unlocks 'mu', re-locks 'mu' when woken up
		}

		fmt.Println("Worker: Condition met, processing data...")
	}()

	// Simulating some work
	time.Sleep(1 * time.Second)

	// Signaler Goroutine
	mu.Lock()
	ready = true  // 3. Update the shared state
	cond.Signal() // 4. Wake up the waiting goroutine (or cond.Broadcast())
	mu.Unlock()

	time.Sleep(500 * time.Millisecond) // Give the worker time to print
}
