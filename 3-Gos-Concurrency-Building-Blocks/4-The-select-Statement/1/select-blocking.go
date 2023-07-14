package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	c := make(chan any)

	go func() {
		time.Sleep(2 * time.Second)
		close(c)
	}()

	fmt.Println("Blocking on read...")
	select {
	case <-c:
		fmt.Printf("Unblocked %v later.\n", time.Since(start))
	}
}
