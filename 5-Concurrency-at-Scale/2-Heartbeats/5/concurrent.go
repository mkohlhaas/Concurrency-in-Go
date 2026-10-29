// Package concurrent provides utilities for concurrent work with
// cancellation support, including a heartbeat channel that signals
// each produced item.
package concurrent

import "time"

func DoWork(done <-chan struct{}, nums ...int) (<-chan struct{}, <-chan int) {
	heartbeat := make(chan struct{}, 1)
	intStream := make(chan int)

	go func() {
		defer close(heartbeat)
		defer close(intStream)

		time.Sleep(2 * time.Second)

		for _, n := range nums {
			select {
			case heartbeat <- struct{}{}:
			default:
			}

			select {
			case <-done:
				return
			case intStream <- n:
			}
		}
	}()

	return heartbeat, intStream
}
