package main

import (
	"fmt"
)

func orDone[T any](done <-chan struct{}, c <-chan T) <-chan T {
	valStream := make(chan T)
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
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

func bridge[T any](done <-chan struct{}, chanStream <-chan <-chan T) <-chan T {
	valStream := make(chan T)

	go func() {
		defer close(valStream)

		for {
			select {
			case <-done:
				return
			case maybeStream, ok := <-chanStream:
				if !ok {
					return // channel chanStream has been closed
				}
				for val := range orDone(done, maybeStream) {
					// Write Select
					select {
					case <-done:
						return
					case valStream <- val:
					}
				}
			}
		}
	}()

	return valStream
}

func main() {
	done := make(chan struct{})
	defer close(done)

	chanStream := func() <-chan <-chan string {
		cs := make(chan (<-chan string))

		go func() {
			defer close(cs)

			// create 10 streams, write a value in it, and write
			// the streams in a stream of streams (the return value)
			for i := range 10 {
				stream := make(chan string, 1)
				stream <- fmt.Sprintf("%d", i)
				close(stream)
				cs <- stream
			}
		}()

		return cs
	}()

	for v := range bridge(done, chanStream) {
		fmt.Printf("%v ", v)
	}

	fmt.Println()
}
