package main

import (
	"fmt"
)

func main() {
	// Create an unbuffered channel of ints and immediately close it.
	// Closing signals that no more values will ever be sent.
	intStream := make(chan int)
	close(intStream)

	// A receive from a closed channel never blocks; it returns at once.
	// The second value, ok, is false because the channel is closed and
	// nothing was ever sent. integer therefore holds the zero value for
	// its type, which is 0 for an int.
	integer, ok := <-intStream

	// Prints "(false): 0" -- ok is false and integer is the zero value.
	fmt.Printf("(%v): %v\n", ok, integer)
}
