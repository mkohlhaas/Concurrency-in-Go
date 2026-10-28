package main

import (
	"fmt"
	"math/rand"
	"time"
)

func repeatFn[T any](done <-chan struct{}, fn func() T) <-chan T {
	valueStream := make(chan T)

	go func() {
		defer close(valueStream)

		for {
			select {
			case <-done:
				return
			case valueStream <- fn():
			}
		}
	}()

	return valueStream
}

func take[T any](done <-chan struct{}, valueStream <-chan T, num int) <-chan T {
	takeStream := make(chan T)

	go func() {
		defer close(takeStream)

		for range num {
			select {
			case <-done:
				return
			case takeStream <- <-valueStream:
			}
		}
	}()

	return takeStream
}

// dummy function
func toInt(done <-chan struct{}, valueStream <-chan int) <-chan int {
	intStream := make(chan int)

	go func() {
		defer close(intStream)

		for v := range valueStream {
			select {
			case <-done:
				return
			case intStream <- v:
			}
		}
	}()

	return intStream
}

// dummy function
func primeFinder(done <-chan struct{}, intStream <-chan int) <-chan int {
	primeStream := make(chan int)

	go func() {
		defer close(primeStream)

		for integer := range intStream {
			integer -= 1
			select {
			case <-done:
				return
			case primeStream <- integer:
			}
		}
	}()

	return primeStream
}

func main() {
	done := make(chan struct{})
	defer close(done)

	start := time.Now()

	randInt := func() int { return rand.Intn(500000000) }

	randIntStream := toInt(done, repeatFn(done, randInt))

	fmt.Println("Primes:")
	for prime := range take(done, primeFinder(done, randIntStream), 10) {
		fmt.Printf("%v\n", prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
}
