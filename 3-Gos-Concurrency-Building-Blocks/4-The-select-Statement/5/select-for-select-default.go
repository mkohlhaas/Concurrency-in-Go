// This program demonstrates using select with a default case to poll a
// "done" channel without blocking. The main loop keeps doing work until a
// watcher goroutine closes the channel after five seconds.
package main

import (
	"fmt"
	"time"
)

func main() {
	// Unbuffered channel used purely as a signal. Closing it broadcasts
	// "done" to every receiver without sending any value.
	done := make(chan struct{})

	fmt.Println("Starting worker; will signal done in 5 seconds...")

	// Watcher goroutine: after a fixed delay, close the channel to tell
	// the main loop to stop.
	go func() {
		fmt.Println("Watcher: sleeping for 5 seconds")
		time.Sleep(5 * time.Second)
		fmt.Println("Watcher: closing done channel")
		close(done)
	}()

	workCounter := 0

	// Label lets us break out of the for loop from inside the select.
loop:
	for {
		// Non-blocking check: if done is closed we take that case,
		// otherwise default runs and the loop continues with work.
		select {
		case <-done:
			fmt.Println("Loop:    received done signal, breaking out of loop")
			break loop
		default:
			fmt.Println("Loop:    done not ready, continuing with work")
		}

		// Simulate one unit of work.
		workCounter++

		fmt.Printf("Loop:    doing work cycle %d\n", workCounter)
		time.Sleep(time.Second)
	}

	fmt.Printf("Achieved %d cycles of work before signalled to stop.\n", workCounter)
}
