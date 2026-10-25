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

func bridge[T any](done <-chan struct{}, chanStream <-chan <-chan T) <-chan T {
	valStream := make(chan T)
	go func() {
		defer close(valStream)
		for {
			var stream <-chan T
			select {
			case <-done:
				return
			case maybeStream, ok := <-chanStream:
				if ok == false {
					return
				}
				stream = maybeStream
			}
			for val := range orDone(done, stream) {
				select {
				case <-done:
					return
				case valStream <- val:
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
			for i := 0; i < 10; i++ {
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
