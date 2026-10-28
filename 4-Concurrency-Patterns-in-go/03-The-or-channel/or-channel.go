package main

import (
	"fmt"
	"time"
)

var orStart = time.Now()

func main() {
	// or: returns a channel that closes when ANY input channel closes.
	var or func(channels ...<-chan any) <-chan any

	or = func(channels ...<-chan any) <-chan any {
		fmt.Printf("[or] called with %d input channel(s) (after %v)\n", len(channels), time.Since(orStart))
		switch len(channels) {
		case 0:
			// Nothing to wait on; nil blocks forever (standard).
			return nil
		case 1:
			// One source: pass through unchanged.
			return channels[0]
		}

		orDone := make(chan any)

		go func() {
			defer close(orDone)

			switch len(channels) {
			case 2:
				// Wait for either of the two.
				var src int
				select {
				case <-channels[0]:
					src = 0
				case <-channels[1]:
					src = 1
				}
				fmt.Printf("[or] source %d closed -> OR done (after %v)\n", src, time.Since(orStart))
			default:
				// 3+: select first 3 directly; fold rest via recursion.
				// orDone is appended so the recursive call terminates.
				fmt.Printf("[or] %d sources; racing the first 3, folding the rest (after %v)\n", len(channels), time.Since(orStart))
				select {
				case <-channels[0]:
					fmt.Printf("[or] source 0 closed -> OR done\n")
				case <-channels[1]:
					fmt.Printf("[or] source 1 closed -> OR done\n")
				case <-channels[2]:
					fmt.Printf("[or] source 2 closed -> OR done\n")
				case <-or(append(channels[3:], orDone)...):
					fmt.Printf("[or] recursive OR resolved -> OR done\n")
				}
			}
		}()

		return orDone
	}

	// sig: creates a channel that closes after the given delay.
	sig := func(name string, after time.Duration) <-chan any {
		fmt.Printf("[sig] %q armed (will close after %v)\n", name, after)
		c := make(chan any)
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	fmt.Println("[main] OR over 5 timer signals; waiting for the first to finish...")
	<-or(
		sig("task-1", 2*time.Hour),
		sig("task-2", 5*time.Minute),
		sig("task-3", 1*time.Second),
		sig("task-4", 1*time.Hour),
		sig("task-5", 1*time.Minute),
	)
	fmt.Printf("Done after %v.\n", time.Since(start))
}
