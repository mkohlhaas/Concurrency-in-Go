package main

import (
	"fmt"
)

func main() {
	orDone := func(done, c <-chan any) <-chan any {
		valStream := make(chan any)
		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if ok == false {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	// A technique called bridging the channels destructures a channel of channels into a simple channel.
	bridge := func(done <-chan any, chanStream <-chan <-chan any) <-chan any {
		valStream := make(chan any)
		go func() {
			defer close(valStream)
			for {
				var stream <-chan any
				select {
				case maybeStream, ok := <-chanStream:
					if !ok {
						return
					}
					stream = maybeStream
				case <-done:
					return
				}
				for val := range orDone(done, stream) {
					// When the stream we're currently looping over is closed, we break out of this loop
					// performing the reads from this channel, and continue with the next iteration of the
					// loop, selecting channels to read from. This provides us with an unbroken stream of values.
					select {
					case valStream <- val:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	// returns a channel of receive-only channels
	genVals := func() <-chan <-chan any {
		chanStream := make(chan (<-chan any))
		go func() {
			defer close(chanStream)
			// create a channel for every integer (1..10)
			for i := 0; i < 10; i++ {
				stream := make(chan any, 1)
				stream <- i
				close(stream)
				chanStream <- stream
			}
		}()
		return chanStream
	}

	// too lazy to care about the done channel
	for v := range bridge(nil, genVals()) {
		fmt.Printf("%v ", v)
	}
	fmt.Println()
}
