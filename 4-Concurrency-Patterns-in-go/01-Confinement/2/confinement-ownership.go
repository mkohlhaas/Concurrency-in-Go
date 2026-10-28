package main

import (
	"fmt"
)

// main demonstrates lexical confinement via channel ownership.
//
// The rule of thumb is: the goroutine that creates a channel should also be
// the one that writes to it, and the one that writes to it should be the one
// that closes it. Here `results` is created inside chanOwner and is never
// written to by anyone else, so the send end is lexically confined to the
// goroutine started inside chanOwner. The consumer only ever sees the
// read-only `<-chan int`, which the type system prevents from sending on or
// closing.
//
// Because of that ownership discipline, the two goroutines never write to the
// channel at the same time and no mutex is required.
func main() {
	// chanOwner creates the channel, writes to it from a goroutine it owns,
	// and hands back only the receive end. The channel is confined to the
	// scope of this closure: nothing outside can send on it or close it.
	chanOwner := func() <-chan int {
		// Buffered with capacity 5 so the producer can write all five values
		// without blocking even if the consumer has not started yet.
		results := make(chan int, 5)

		// This goroutine is the only writer, which satisfies the ownership
		// rule: one goroutine sends and closes.
		go func() {
			// Only the owner of the send end may close the channel; closing
			// here signals the consumer's range loop that no more values
			// are coming.
			defer close(results)

			for i := range 5 {
				results <- i
			}
		}()

		// Returning a receive-only channel (`<-chan int`) makes the
		// confinement explicit in the type: callers literally cannot send
		// on it or close it.
		return results
	}

	// consumer takes a receive-only channel. It cannot write to or close it,
	// so it cannot violate the ownership the producer established.
	consumer := func(results <-chan int) {
		// range drains the channel until the owner closes it.
		for result := range results {
			fmt.Printf("Received: %d\n", result)
		}
		fmt.Println("Done receiving!")
	}

	// Hand-off: the producer goroutine is already writing, and the consumer
	// reads the same channel through the restricted read-only view.
	results := chanOwner()
	consumer(results)
}
