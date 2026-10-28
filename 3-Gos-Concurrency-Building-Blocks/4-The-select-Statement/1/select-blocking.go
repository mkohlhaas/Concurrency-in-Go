package main

import (
	"fmt"
	"time"
)

func main() {
	// Record the start time so we can measure how long the receive blocks.
	start := time.Now()

	// done is a signal-only channel: it carries no data, only the event of
	// being closed. Because it is unbuffered, a send would block until a
	// receiver is ready, but closing it releases all current receivers.
	done := make(chan struct{})

	// Simulate work in the background. After two seconds the goroutine
	// closes done, which unblocks the select below.
	go func() {
		time.Sleep(2 * time.Second)
		close(done)
	}()

	// This prints before the select starts waiting.
	fmt.Println("Blocking on read...")

	// select waits until one of its cases can proceed. There is only one
	// case here, so this blocks until done is closed. A plain <-done would
	// behave the same; the book uses select to introduce the statement.
	select {
	case <-done:
		// This runs only after the goroutine closes done, roughly two
		// seconds after start.
		fmt.Printf("Unblocked %v later.\n", time.Since(start))
	}
}
