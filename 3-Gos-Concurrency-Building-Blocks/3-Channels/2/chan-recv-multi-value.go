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

	// A receive can return a second, boolean value: the "ok" idiom.
	// Here ok is true, meaning salutation holds a value that was actually
	// sent on the channel. It is false only when receiving from a channel
	// that has been closed, in which case salutation is the zero value.
	salutation, ok := <-stringStream
	fmt.Printf("(%v): %v\n", ok, salutation)
}
