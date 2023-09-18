package main

import (
	"fmt"
)

func main() {
	c1 := make(chan any)
	c2 := make(chan any)
	close(c1)
	close(c2)

	var c1Count, c2Count int
	for i := 1000; i > 0; i-- {
		select {
		case <-c1:
			c1Count++
		case <-c2:
			c2Count++
		}
	}

	fmt.Printf("c1 count: %d\nc2 count: %d\n", c1Count, c2Count)
}
