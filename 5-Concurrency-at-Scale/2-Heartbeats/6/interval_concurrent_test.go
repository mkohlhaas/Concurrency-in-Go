package main

import (
	"testing"
	"time"
)

func DoWork(done <-chan any, pulseInterval time.Duration, nums ...int) (<-chan any, <-chan int) {
	heartbeat := make(chan any, 1)
	intStream := make(chan int)
	go func() {
		defer close(heartbeat)
		defer close(intStream)

		time.Sleep(2 * time.Second)

		pulse := time.Tick(pulseInterval)
	numLoop:
		for _, n := range nums {
			for {
				select {
				case <-done:
					return
				case <-pulse:
					select {
					case heartbeat <- struct{}{}:
					default:
					}
				case intStream <- n:
					continue numLoop // next value in nums
				}
			}
		}
	}()

	return heartbeat, intStream
}

func TestDoWork_GeneratesAllNumbers(t *testing.T) {
	done := make(chan any)
	defer close(done)

	intSlice := []int{0, 1, 2, 3, 5}
	const timeout = 2 * time.Second
	heartbeat, results := DoWork(done, timeout/2, intSlice...)

	<-heartbeat

	i := 0
	for {
		select {
		case r, ok := <-results:
			if ok == false {
				return
			} else if expected := intSlice[i]; r != expected {
				t.Errorf("index %v: expected %v, but received %v,", i, expected, r)
			}
			i++
		case <-heartbeat:
		case <-time.After(timeout):
			t.Fatal("test timed out")
		}
	}
}
