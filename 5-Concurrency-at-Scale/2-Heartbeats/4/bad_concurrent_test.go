package main_test

import (
	"testing"
	"time"
)

func DoWork(
	done <-chan struct{}, nums ...int,
) (<-chan struct{}, <-chan int) {
	heartbeat := make(chan struct{}, 1)
	intStream := make(chan int)

	go func() {
		defer close(heartbeat)
		defer close(intStream)

		// Simulating a producer that is alive but slow to start.
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
	done := make(chan struct{})
	defer close(done)

	intSlice := []int{0, 1, 2, 3, 5}
	heartbeat, results := DoWork(done, intSlice...)

	// Wait for the first heartbeat: it proves the producer is alive
	// despite the initial work delay.
	<-heartbeat

	i := 0
	for r := range results {
		if expected := intSlice[i]; r != expected {
			t.Errorf("index %v: expected %v, but received %v,", i, expected, r)
		}
		i++
	}
}

func BenchmarkDoWork(b *testing.B) {
	nums := []int{0, 1, 2, 3, 5}

	for b.Loop() {
		done := make(chan struct{})
		_, results := DoWork(done, nums...)

		for range results {
		}
		close(done)
	}
}
