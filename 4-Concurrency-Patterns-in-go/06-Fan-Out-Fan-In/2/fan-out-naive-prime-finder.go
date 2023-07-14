package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func main() {
	repeatFn := func(done <-chan any, fn func() any) <-chan any {
		valueStream := make(chan any)

		go func() {
			defer close(valueStream)
      defer fmt.Println("repeatFn is shutting down. (defer)")
			for {
				select {
				case <-done:
					fmt.Println("repeatFn is shutting down.")
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
      defer fmt.Println("takeStream is shutting down. (defer)")
			for i := 0; i < num; i++ {
				select {
				case <-done:
					fmt.Println("take is shutting down.")
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
      defer fmt.Println("intStream is shutting down. (defer)")
			for v := range valueStream {
				select {
				case <-done:
					fmt.Println("toInt is shutting down.")
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
      defer fmt.Println("primeFinder is shutting down. (defer)")
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
						fmt.Println("primeFinder is shutting down.")
						return
					case primeStream <- integer:
					}
				}
			}
		}()

		return primeStream
	}

	fanIn := func(done <-chan any, channels ...<-chan any) <-chan any {
		var wg sync.WaitGroup
		multiplexedStream := make(chan any)

		multiplex := func(c <-chan any) {
			defer wg.Done()
      defer fmt.Println("fanIn is shutting down. (defer)")
			for i := range c {
				select {
				case <-done:
					fmt.Println("fanIn is shutting down.")
					return
				case multiplexedStream <- i:
				}
			}
		}

		// select from all the channels
		wg.Add(len(channels))
		for _, c := range channels {
			go multiplex(c)
		}

		// wait for all the reads to complete
		go func() {
			wg.Wait()
			close(multiplexedStream)
		}()

		return multiplexedStream
	}

	done := make(chan any)
	// defer close(done)

	start := time.Now()

	rand := func() any { return rand.Intn(50000000) }

	randIntStream := toInt(done, repeatFn(done, rand))

	numFinders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", numFinders)
	finders := make([]<-chan any, numFinders)
	fmt.Println("Primes:")
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}

	for prime := range take(done, fanIn(done, finders...), 10) {
		fmt.Printf("  %d\n", prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
	close(done)
  time.Sleep(1 * time.Second)
}
