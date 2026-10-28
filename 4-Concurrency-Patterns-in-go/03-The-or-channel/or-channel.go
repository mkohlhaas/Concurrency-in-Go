package main

import (
	"fmt"
	"time"
)

func main() {
	// or: returns a channel that closes when ANY input channel closes.
	var or func(channels ...<-chan any) <-chan any

	or = func(channels ...<-chan any) <-chan any {
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
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				// 3+: select first 3 directly; fold rest via recursion.
				// orDone is appended so the recursive call terminates.
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()

		return orDone
	}

	// sig: creates a channel that closes after the given delay.
	sig := func(after time.Duration) <-chan any {
		c := make(chan any)
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("Done after %v.\n", time.Since(start))
}
