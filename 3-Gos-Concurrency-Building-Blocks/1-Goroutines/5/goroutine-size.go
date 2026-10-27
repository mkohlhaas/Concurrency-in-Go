package main

import (
	"fmt"
	"runtime"
	"sync"
)

const numGoroutines = 1e6

func main() {
	memConsumed := func() uint64 {
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		return s.Sys
	}

	var c <-chan struct{}
	var wg sync.WaitGroup
	// Done must be called before blocking on c: the goroutines have to stay
	// parked on a nil channel while memAfter is measured. They never finish --
	// this is a deliberate leak, see README.md.
	noop := func() {
		wg.Done()
		<-c
	}

	memBefore := memConsumed()

	wg.Add(int(numGoroutines))
	for range int(numGoroutines) {
		go noop()
	}

	wg.Wait()

	memAfter := memConsumed()

	fmt.Printf("%.3fKB\n", float64(memAfter-memBefore)/numGoroutines/1024)
}
