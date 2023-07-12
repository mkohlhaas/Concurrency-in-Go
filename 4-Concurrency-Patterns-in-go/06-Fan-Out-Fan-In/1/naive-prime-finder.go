package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	repeatFn := func(done <-chan any, fn func() any) <-chan any {
		valueStream := make(chan any)

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

	take := func(done <-chan any, valueStream <-chan any, num int) <-chan any {
		takeStream := make(chan any)

		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}
		}()

		return takeStream
	}

	toInt := func(done <-chan any, valueStream <-chan any) <-chan int {
		intStream := make(chan int)

		go func() {
			defer close(intStream)
			for v := range valueStream {
				select {
				case <-done:
					return
				case intStream <- v.(int):
				}
			}
		}()

		return intStream
	}

	primeFinder := func(done <-chan any, intStream <-chan int) <-chan any {
		primeStream := make(chan any)

		go func() {
			defer close(primeStream)
			for integer := range intStream {
				integer -= 1
				prime := true
				for divisor := integer - 1; divisor > 1; divisor-- {
					if integer%divisor == 0 {
						prime = false
						break
					}
				}

				if prime {
					select {
					case <-done:
						return
					case primeStream <- integer:
					}
				}
			}
		}()

		return primeStream
	}

	rand := func() any { return rand.Intn(50000000) }

	done := make(chan any)
	defer close(done)

	start := time.Now()

	randIntStream := toInt(done, repeatFn(done, rand))
	fmt.Println("Primes:")
	for prime := range take(done, primeFinder(done, randIntStream), 10) {
		fmt.Printf("  %d\n", prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
}
