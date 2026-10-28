package main

import (
	"fmt"
	"sync"
)

func main() {
	// begin is a signal-only channel: it carries no data, only the event
	// of being closed. Receiving from it blocks until that happens.
	begin := make(chan struct{})

	// wg tracks how many goroutines are still running so main can wait
	// for all of them before exiting.
	var wg sync.WaitGroup

	for i := range 5 {
		// wg.Go adds 1 to the counter, starts the goroutine, and calls
		// wg.Done() when the function returns, so no manual Add/Done
		// bookkeeping is needed. Since Go 1.22 each iteration has its own
		// i, so the closure captures the right value without a parameter.
		wg.Go(func() {
			// Block here until main closes begin below.
			<-begin
			fmt.Printf("%v has begun\n", i)
		})
	}

	fmt.Println("Unblocking goroutines...")

	// Closing the channel unblocks all five receivers at once (a "close"
	// broadcasts, while a send would only release one goroutine). Reading
	// from a closed channel returns immediately with the zero value.
	close(begin)

	// Wait for every goroutine to print before main returns and the
	// program exits.
	wg.Wait()
}
