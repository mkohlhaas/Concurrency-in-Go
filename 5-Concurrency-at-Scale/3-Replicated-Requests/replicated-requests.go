package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	doWork := func(done <-chan any, id int, wg *sync.WaitGroup, result chan<- int) {
		started := time.Now()
		defer wg.Done()

		// Simulate random load
		simulatedLoadTime := time.Duration(1+rand.Intn(5)) * time.Second
		select {
		case <-done:
		case <-time.After(simulatedLoadTime):
		}

		// send result (= goroutine id)
		select {
		case <-done:
		case result <- id:
		}

		// Display how long handlers would have taken
		took := max(time.Since(started), simulatedLoadTime)
		fmt.Printf("goroutine %v took %v\n", id, took)
	}

	done := make(chan any)
	results := make(chan int)
	var wg sync.WaitGroup
	wg.Add(10)

	// Start 10 goroutines using the same results channel.
	// All goroutines will finish.
	for i := range 10 {
		go doWork(done, i, &wg, results)
	}

	firstResult := <-results
	close(done)

	wg.Wait()
	fmt.Printf("Received first answer from #%v.\n", firstResult)
}
