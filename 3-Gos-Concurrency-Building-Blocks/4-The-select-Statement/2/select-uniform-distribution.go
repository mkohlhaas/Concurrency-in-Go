package main

import (
	"fmt"
)

func main() {
	// Two signal-only channels. They carry no data, so struct{} is enough.
	c1 := make(chan struct{})
	c2 := make(chan struct{})

	// Closing a channel makes every receive from it succeed immediately
	// with the zero value. Both cases below are therefore ready at all
	// times, which is exactly what this example needs.
	close(c1)
	close(c2)

	// Counters for how often each case is chosen.
	var c1Count, c2Count int

	// Run the select many times so the distribution becomes visible.
	for range 1000 {
		// Both cases are always ready, so select never blocks. Instead of
		// evaluating the cases top to bottom, the runtime picks a ready
		// case uniformly at random. This prevents the first case from
		// starving the others and yields roughly even counts (~500 each).
		select {
		case <-c1:
			c1Count++
		case <-c2:
			c2Count++
		}
	}

	fmt.Printf("c1 count: %d\nc2 count: %d\n", c1Count, c2Count)
}
