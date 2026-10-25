package main

import (
	"fmt"
	"math/rand"
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

func main() {
	done := make(chan struct{})
	defer close(done)

	randInt := func() int { return rand.Intn(100) }

	for num := range take(done, repeatFn(done, randInt), 10) {
		fmt.Println(num)
	}
}
