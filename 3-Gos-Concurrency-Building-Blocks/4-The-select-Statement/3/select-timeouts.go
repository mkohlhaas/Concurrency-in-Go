package main

import (
	"fmt"
	"time"
)

func main() {
	// A nil channel that only receives. Receiving from a nil channel blocks
	// forever, so this case can never become ready. It stands in for a
	// channel that might never deliver a value.
	var c <-chan int

	// select waits for whichever case becomes ready first. Since the receive
	// from c never completes, the timeout case is the one that wins.
	select {
	case <-c:
		// Unreachable in this example: a receive from a nil channel blocks
		// forever, so this branch never runs.
	case <-time.After(time.Second):
		// time.After returns a channel that receives a single value after
		// the given duration. This turns "wait at most one second" into an
		// ordinary select case, which is the idiomatic way to add timeouts.
		fmt.Println("Timed out after a second.")
	}
}
