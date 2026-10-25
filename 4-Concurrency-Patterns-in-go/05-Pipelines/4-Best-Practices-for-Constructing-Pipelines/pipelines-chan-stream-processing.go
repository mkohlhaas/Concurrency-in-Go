package main

import (
	"fmt"
)

func main() {
	generator := func[T any](done <-chan struct{}, integers ...T) <-chan T {
		intStream := make(chan T, len(integers))

		go func() {
			defer close(intStream)
			for _, i := range integers {
				select {
				case <-done:
					return
				case intStream <- i:
				}
			}
		}()

		return intStream
	}

	multiply := func[T ~int | ~int64 | ~float64](done <-chan struct{}, intStream <-chan T, multiplier T) <-chan T {
		multipliedStream := make(chan T)

		go func() {
			defer close(multipliedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case multipliedStream <- i * multiplier:
				}
			}
		}()

		return multipliedStream
	}

	add := func[T ~int | ~int64 | ~float64](done <-chan struct{}, intStream <-chan T, additive T) <-chan T {
		addedStream := make(chan T)

		go func() {
			defer close(addedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case addedStream <- i + additive:
				}
			}
		}()

		return addedStream
	}

	done := make(chan struct{})
	defer close(done)

	intStream := generator(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, intStream, 2), 1), 2)

	for v := range pipeline {
		fmt.Println(v)
	}
}
