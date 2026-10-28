package main

import (
	"fmt"
	"sync"
	"time"
)

// Or takes a variadic list of channels and returns a single new channel.
// The returned channel receives the first value that becomes ready on any
// of the input channels — the concurrency equivalent of boolean ||.
func Or(channels ...<-chan any) <-chan any {
	out := make(chan any, 1)

	if len(channels) == 0 {
		close(out)
		return out
	}

	done := make(chan struct{})
	var once sync.Once

	for _, ch := range channels {
		go func(c <-chan any) {
			select {
			case v, ok := <-c:
				once.Do(func() {
					if ok {
						out <- v
					}
					close(done)
				})
			case <-done:
			}
		}(ch)
	}

	go func() {
		<-done
		close(out)
	}()

	return out
}

// sig: creates a channel that closes after the given delay.
func sig(after time.Duration) <-chan any {
	c := make(chan any)

	go func() {
		defer close(c)
		time.Sleep(after)
	}()

	return c
}

func main() {
	start := time.Now()
	<-Or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("Done after %v.\n", time.Since(start))
}
