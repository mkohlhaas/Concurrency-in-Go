// This example demonstrates starvation. Both workers do the same amount of
// work per round (3 ns), but the greedy worker holds the mutex for longer
// contiguous stretches, so it reacquires it before the polite worker gets a
// chance and completes many more work loops in the same wall-clock time.
//
// Note: time.Sleep rounds up to the timer granularity, so the real sleeps are
// much longer than the named durations. The ratio between the workers is what
// matters here, not the absolute timings.
//
// Since Go 1.9, sync.Mutex has a starvation mode: after waiting for ~1 ms, a
// waiter is handed the lock directly, so the greedy worker cannot starve the
// polite one indefinitely. That is why the greedy worker only wins by a
// fraction (roughly 35% in practice) rather than by many multiples as in the
// original example.

package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	var sharedLock sync.Mutex
	const runtime = 1 * time.Second

	// The greedy worker does 3 ns of work per lock acquisition, so it holds the
	// lock in longer contiguous stretches and acquires it far less often than
	// the polite worker. That is what lets it starve the polite worker.
	greedyWorker := func() {
		defer wg.Done()

		var count int

		// Run for 1 second, doing a nominally 3 ns chunk of work per round.
		for begin := time.Now(); time.Since(begin) <= runtime; {
			sharedLock.Lock()
			time.Sleep(3 * time.Nanosecond) // a 3 ns chunk of work, done while holding the lock
			sharedLock.Unlock()
			count++
		}

		fmt.Printf("Greedy worker: %v work loops\n", count)
	}

	// The polite worker locks and unlocks often, releasing the lock between each
	// chunk of work so other goroutines can make progress.
	politeWorker := func() {
		defer wg.Done()

		var count int

		// Run for 1 second, doing three nominally 1 ns chunks of work per round.
		for begin := time.Now(); time.Since(begin) <= runtime; {
			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond) // a 1 ns chunk of work
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond) // a 1 ns chunk of work
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond) // a 1 ns chunk of work
			sharedLock.Unlock()

			count++
		}

		fmt.Printf("Polite worker: %v work loops\n", count)
	}

	wg.Add(2)
	go greedyWorker()
	go politeWorker()

	wg.Wait()
}
