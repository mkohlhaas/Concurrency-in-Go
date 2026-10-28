package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
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

func fanIn[T any](done <-chan struct{}, channels ...<-chan T) <-chan T {
	multiplexedStream := make(chan T)
	var wg sync.WaitGroup

	multiplex := func(c <-chan T) {
		defer wg.Done()

		for i := range c {
			select {
			case <-done:
				return
			case multiplexedStream <- i:
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}

func main() {
	done := make(chan struct{})
	defer close(done)

	start := time.Now()

	randInt := func() int { return rand.Intn(500000000) }

	// start prime finders
	randIntStream := toInt(done, repeatFn(done, randInt))
	numFinders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", numFinders)

	finders := make([]<-chan int, numFinders)
	for i := range numFinders {
		finders[i] = primeFinder(done, randIntStream)
	}

	// print prime numbers
	const numPrimes = 10
	fmt.Println("Primes:")
	for prime := range take(done, fanIn(done, finders...), numPrimes) {
		fmt.Printf("%v\n", prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
}
