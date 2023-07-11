package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan any)

	go func() {
		time.Sleep(5 * time.Second)
		close(done)
	}()

	workCounter := 0

loop:
	for {
		select {
		case <-done:
			break loop
		default:
		}

		// Simulate work
		workCounter++
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("Achieved %v cycles of work before signalled to stop.\n", workCounter)
}
