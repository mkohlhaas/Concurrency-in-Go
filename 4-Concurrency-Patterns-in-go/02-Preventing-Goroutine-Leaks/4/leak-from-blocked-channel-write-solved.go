package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	newRandStream := func(done <-chan struct{}) <-chan int {
		randStream := make(chan int)

		go func() {
			defer fmt.Println("newRandStream closure exited.")
			defer close(randStream)

			for {
				select {
				case randStream <- rand.Int():
				case <-done:
					return
				}
			}
		}()

		return randStream
	}

	done := make(chan struct{})
	randStream := newRandStream(done)
	fmt.Println("5 random ints:")
	for i := range 5 {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}

	// shutdown the random stream goroutine (no leak)
	close(done)

	// Simulate ongoing work
	time.Sleep(time.Second)

	fmt.Println("Done!")
}
