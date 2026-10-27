package main

import (
	"sync"
	"testing"
)

func Benchmark_ContextSwitch(b *testing.B) {
	var wg sync.WaitGroup
	var token struct{}

	startNow := make(chan struct{})
	c := make(chan struct{})

	sender := func() {
		<-startNow
		for range b.N {
			c <- token
		}
	}

	receiver := func() {
		<-startNow
		for range b.N {
			<-c
		}
	}

	wg.Go(sender)
	wg.Go(receiver)

	b.StartTimer()
	close(startNow)

	wg.Wait()
}
