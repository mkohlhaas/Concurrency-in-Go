package main

import (
	"fmt"
	"time"
)

func main() {
	doWork := func(done <-chan struct{}, strings <-chan string) <-chan struct{} {
		terminated := make(chan struct{})

		go func() {
			defer fmt.Println("doWork exited.")
			defer close(terminated)

			for {
				select {
				case s := <-strings:
					// Do something interesting
					fmt.Println(s)
				case <-done:
					return
				}
			}
		}()

		return terminated
	}

	done := make(chan struct{})
	strings := make(chan string)
	terminated := doWork(done, strings)

	// using the strings channel
	go func() {
		// Send a string to doWork every 200 milliseconds.
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for i := 1; ; i++ {
			select {
			case <-ticker.C:
				select {
				case strings <- fmt.Sprintf("string %d", i):
				case <-done:
					return
				}
			case <-done:
				return
			}
		}
	}()

	// using the done channel
	go func() {
		// Cancel the operation after 1 second.
		time.Sleep(time.Second)
		fmt.Println("Canceling doWork goroutine...")
		close(done)
	}()

	<-terminated

	fmt.Println("Done.")
}
