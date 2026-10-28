package main

import (
	"fmt"
	"math/rand"
)

func main() {
	newRandStream := func() <-chan int {
		randStream := make(chan int)

		go func() {
			defer fmt.Println("newRandStream closure exited.")
			defer close(randStream)

			for {
				randStream <- rand.Int()
			}
		}()

		return randStream
	}

	randStream := newRandStream()

	fmt.Println("5 random ints:")
	for i := range 5 {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}

	// randStream is never shut down

	// for i := range randStream {
	// 	fmt.Printf("%d\n", i)
	// }
}
