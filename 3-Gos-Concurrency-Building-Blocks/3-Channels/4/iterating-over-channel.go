package main

import (
	"fmt"
)

func main() {
	// An unbuffered channel of ints. Sends block until a receiver is ready,
	// and receives block until a sender is ready.
	intStream := make(chan int)

	// The producer runs concurrently so that it can hand values to the
	// consumer below without deadlocking on an unbuffered channel.
	go func() {
		// Closing the channel signals "no more values will be sent". This is
		// what lets the range loop in main terminate instead of blocking.
		defer close(intStream)

		// range over an integer yields 0, 1, 2, 3, 4 (Go 1.22+).
		for i := range 5 {
			intStream <- i
		}
	}()

	// Ranging over a channel receives values until the channel is closed.
	// The loop ends once intStream is closed and all sent values are drained.
	for i := range intStream {
		fmt.Printf("%v ", i)
	}

	// Finish the output line with a newline.
	fmt.Println()
}
