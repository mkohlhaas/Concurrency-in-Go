// This program builds a bounded queue out of sync.Cond: main acts as the
// producer, and every enqueue spawns a consumer that removes an item one
// second later. The producer blocks whenever the queue grows past maxQueued,
// which forces the two sides to take turns.
package main

import (
	"fmt"
	"sync"
	"time"
)

const capacity = 20

// maxQueued is the high-water mark that makes producers wait for a consumer.
// The queue can transiently hold maxQueued+1 items, because the producer only
// blocks while len(queue) > maxQueued.
const maxQueued = 2

func main() {
	// The elements carry no data; only len(queue) matters. struct{} is a
	// zero-size type, so the buffer stores no per-item payload.
	queue := make([]struct{}, 0, capacity)

	// cond.L is the *sync.Mutex guarding queue. Producer and consumers share
	// this one mutex, which is what makes Wait and Signal cooperate.
	cond := sync.NewCond(&sync.Mutex{})

	// Tracks in-flight consumers so main does not exit before they finish.
	var consumers sync.WaitGroup

	// Running counts for the log lines. Both are guarded by cond.L: the
	// producer increments added while holding the lock, and every consumer
	// increments removed while holding it, so no atomics are needed.
	var added, removed int

	// Consumer: wait, then remove one item and wake the producer.
	removeFromQueue := func(delay time.Duration) {
		defer consumers.Done()

		// Sleep outside the critical section so the producer is not blocked.
		time.Sleep(delay)

		cond.L.Lock()
		defer cond.L.Unlock()

		queue = queue[1:]
		removed++
		fmt.Printf("Removed from queue: #%d\n", removed)

		// Signal wakes exactly one waiter (only the producer waits here).
		// Signal while still holding the lock, as sync.Cond expects.
		cond.Signal()
	}

	// Producer: enqueue capacity items, one per iteration.
	for range capacity {
		cond.L.Lock()

		// Always re-check the predicate in a loop: after Wait returns, another
		// goroutine may have changed the queue again (and wake-ups can be
		// spurious). Wait atomically unlocks cond.L while parked and re-locks
		// it before returning.
		for len(queue) > maxQueued {
			cond.Wait()
		}

		queue = append(queue, struct{}{})
		added++
		fmt.Printf("Adding to queue: #%d\n", added)

		// Register before starting the goroutine, so Wait can never miss it.
		consumers.Add(1)
		go removeFromQueue(time.Second)

		cond.L.Unlock()
	}

	// Wait outside the lock: a consumer needs cond.L before it can finish.
	consumers.Wait()
}
