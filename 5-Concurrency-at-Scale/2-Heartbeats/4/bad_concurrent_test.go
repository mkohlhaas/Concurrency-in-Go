package main

import (
	"testing"
	"time"
)

func DoWork(
	done <-chan any, nums ...int) (<-chan any, <-chan int) {
	heartbeat := make(chan any, 1)
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
func TestDoWork_GeneratesAllNumbers(t *testing.T) {
	done := make(chan any)
	defer close(done)

	intSlice := []int{0, 1, 2, 3, 5}
	_, results := DoWork(done, intSlice...)

	for i, expected := range intSlice {
		select {
		case r := <-results:
			if r != expected {
				t.Errorf("index %v: expected %v, but received %v,", i, expected, r)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("test timed out")
		}
	}
}
