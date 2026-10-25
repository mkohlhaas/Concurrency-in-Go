package main

import "fmt"

func main() {
	doWork := func(strings <-chan string) <-chan struct{} {
		completed := make(chan struct{})

		go func() {
			defer fmt.Println("doWork exited.")
			defer close(completed)
			for s := range strings {
				// Do something interesting
				fmt.Println(s)
			}
		}()

		return completed
	}

	doWork(nil)
	// Perhaps more work is done here
	fmt.Println("Done.")
}
