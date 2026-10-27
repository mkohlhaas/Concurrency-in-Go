// This program benchmarks sync.RWMutex read locking against plain exclusive
// locking as the number of concurrent readers grows. It is adapted from
// "Concurrency in Go" (ch. 3); see README.md for the interpretation of the
// results.
package main

import (
	"fmt"
	"math"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

func main() {
	// producer simulates an occasional writer: it takes the full (write) lock
	// five times and releases it each time. l must therefore be a write-side
	// locker, not an RWMutex's read-side RLocker.
	producer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		for range 5 {
			l.Lock()
			// Sleep(0) is a scheduler yield, not a real pause. It keeps the
			// critical section non-empty without dominating the runtime.
			time.Sleep(0)
			l.Unlock()
		}
	}

	// observer simulates a reader: it acquires the lock once and releases it.
	// When l is an RWMutex's RLocker this takes the shared read lock; when it
	// is the RWMutex itself this takes the exclusive lock.
	observer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		l.Lock()
		time.Sleep(0)
		l.Unlock()
	}

	// test runs one producer and count observers concurrently, waits for all of
	// them to finish, and returns the elapsed wall-clock time. The producer uses
	// mutex and the observers use rwMutex, so callers can pair an exclusive
	// writer with either read-locked or exclusively locked readers.
	test := func(count int, mutex, rwMutex sync.Locker) time.Duration {
		var wg sync.WaitGroup

		wg.Add(count + 1) // one producer plus count observers
		beginTestTime := time.Now()
		go producer(&wg, mutex)

		for range count {
			go observer(&wg, rwMutex)
		}

		wg.Wait()
		return time.Since(beginTestTime)
	}

	// tabwriter keeps the columns of the results table aligned.
	tw := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
	defer func() { _ = tw.Flush() }()

	// m is the single mutex behind both variants below: &m is its write side and
	// readerLock is its read side. Both views refer to the same underlying
	// mutex, so the two variants are directly comparable.
	var m sync.RWMutex
	readerLock := m.RLocker()

	// Each test call blocks until all of its goroutines finish, so the two
	// variants (and successive rows) never use m concurrently.
	_, _ = fmt.Fprintf(tw, "Readers\tRWMutex (read lock)\tMutex (exclusive)\n")
	// Double the reader count on each row: 1, 2, 4, ..., 2^19 = 524288.
	for i := range 20 {
		count := int(math.Pow(2, float64(i)))
		_, _ = fmt.Fprintf(
			tw,
			"%d\t%v\t%v\n",
			count,
			// Read-lock variant: the writer takes &m's write lock while readers
			// take readerLock's shared read lock.
			test(count, &m, readerLock),
			// Baseline: producer and observers all take &m's write lock, so the
			// RWMutex behaves like a plain sync.Mutex.
			test(count, &m, &m),
		)
	}
}
