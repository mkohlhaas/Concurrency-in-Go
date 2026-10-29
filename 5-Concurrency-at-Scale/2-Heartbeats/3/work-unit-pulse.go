package main

import (
	"fmt"
	"math/rand"
)

func main() {
	doWork := func(done <-chan struct{}) (<-chan struct{}, <-chan int) {
		heartbeatStream := make(chan struct{}, 1)
		workStream := make(chan int)

		go func() {
			defer close(heartbeatStream)
			defer close(workStream)

			for range 10 {
				select {
				case heartbeatStream <- struct{}{}:
				default: // consumer hasn't read previous pulse; drop this one rather than block
				}

				select {
				case <-done:
					return
				case workStream <- rand.Intn(10):
				}
			}
		}()

		return heartbeatStream, workStream
	}

	done := make(chan struct{})
	defer close(done)

	heartbeat, results := doWork(done)

	for {
		select {
		case _, ok := <-heartbeat:
			if ok {
				fmt.Println("pulse")
			} else {
				return // channel closed by producer's defer
			}
		case r, ok := <-results:
			if ok {
				fmt.Printf("results %v\n", r)
			} else {
				return // channel closed by producer's defer
			}
		}
	}
}
