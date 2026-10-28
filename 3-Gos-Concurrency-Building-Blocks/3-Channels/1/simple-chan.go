package main

import (
	"fmt"
)

func main() {
	// Create an unbuffered channel of strings. Because it is unbuffered,
	// a send only completes once a receiver is ready to take the value.
	stringStream := make(chan string)

	// Start a goroutine that sends a single value into the channel.
	// This must run concurrently: an unbuffered send blocks until
	// something receives from the channel.
	go func() {
		stringStream <- "Hello channels!"
	}()

	// Receive the value from the channel and print it. This receive
	// unblocks the goroutine above, which can then finish.
	fmt.Println(<-stringStream)
}
