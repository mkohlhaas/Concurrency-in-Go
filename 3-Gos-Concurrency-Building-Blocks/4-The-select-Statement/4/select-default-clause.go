package main

import (
	"fmt"
	"time"
)

func main() {
	// Record the start time so we can measure how long the select takes.
	start := time.Now()

	// Declare two nil receive-only channels. A receive from a nil channel
	// blocks forever, so neither of these cases is ever ready.
	var c1, c2 <-chan int

	// select blocks until one of its cases can proceed. If no case is ready,
	// the default clause runs immediately instead of blocking, which makes
	// this select a non-blocking operation.
	select {
	case <-c1:
		// Never reached: c1 is nil, so this case is never ready.
	case <-c2:
		// Never reached: c2 is nil, so this case is never ready.
	default:
		// Always reached, since both channels are nil. Prints a duration of
		// approximately zero because select does not wait here.
		fmt.Printf("In default after %v.\n", time.Since(start))
	}
}
