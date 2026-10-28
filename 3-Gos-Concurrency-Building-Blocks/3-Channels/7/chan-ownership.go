package main

import (
	"fmt"
)

func main() {
	// This function demonstrates the channel ownership pattern: the
	// goroutine that creates a channel owns it, and it is the only one
	// allowed to close that channel. Here chanOwner both creates
	// resultStream and closes it, while callers only ever receive from it.
	//
	// Note the return type: a receive-only channel (<-chan int). Because the
	// return type has no send direction, callers physically cannot send on
	// or close the channel, so the ownership rules are enforced by the
	// type system rather than by convention alone.
	chanOwner := func() <-chan int {
		// A buffered channel with a capacity of 5. The buffer decouples the
		// sender from the receiver: the goroutine below can place up to five
		// values in the channel before it has to wait for someone to receive.
		resultStream := make(chan int, 5)

		// Spawn the producing goroutine. Ownership of the write end stays in
		// this function through this closure, which is why the close happens
		// here as well.
		go func() {
			// Deferring the close guarantees the channel is closed exactly
			// once, even if the loop were to panic or return early. Closing
			// signals to receivers that no more values are coming; ranging
			// over a still-open channel would block forever instead of ending.
			defer close(resultStream)

			// Send the values 0 through 4 into the channel. Ranging over an
			// integer yields 0, 1, ..., n-1 (Go 1.22+).
			for i := range 5 {
				resultStream <- i
			}
		}()

		// Hand back the read-only view of the channel. The goroutine started
		// above keeps running after this function returns; the caller simply
		// waits for values to arrive.
		return resultStream
	}

	// The caller only holds a receive-only channel. It has no way to close
	// or send on it, so no other part of this program can violate ownership.
	resultStream := chanOwner()

	// Range over the channel until it is closed. Once the producer's deferred
	// close runs, the range ends cleanly; otherwise this loop would block
	// forever waiting for a value that never arrives.
	for result := range resultStream {
		fmt.Printf("Received: %d\n", result)
	}

	fmt.Println("Done receiving!")
}
