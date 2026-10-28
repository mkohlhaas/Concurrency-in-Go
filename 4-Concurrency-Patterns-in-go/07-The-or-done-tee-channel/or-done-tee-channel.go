package main

import "fmt"

func repeat[T any](done <-chan struct{}, values ...T) <-chan T {
	valueStream := make(chan T)

	go func() {
		defer close(valueStream)

		for {
			for _, v := range values {
				// Pre-emptive check: Exit immediately if done,
				// before even risking a random selection below
				select {
				case <-done:
					return
				default:
				}

				// Unintended Progress
				//
				// Because Go picks randomly when multiple cases are ready, here is the exact scenario where the flaw exposes itself:
				// 1. The parent goroutine closes done because it wants to stop everything.
				// 2. At the exact same microsecond, a consumer reads from valueStream.
				// 3. Because both cases are ready, Go's runtime randomly chooses case valueStream <- v.
				// 4. The loop successfully sends the value, moves to the next item in the slice, and continues forward.
				// So it doesn't stop immediately. There the select above to kill the goroutine immediately.
				select {
				case <-done:
					return
				case valueStream <- v:
				}
			}
		}
	}()

	return valueStream
}

func take[T any](done <-chan struct{}, valueStream <-chan T, num int) <-chan T {
	takeStream := make(chan T)
	go func() {
		defer close(takeStream)

		for range num {
			select {
			case <-done:
				return
			case takeStream <- <-valueStream:
			}
		}
	}()

	return takeStream
}

// orDone() is just a pass-through layer. It acts like a smart funnel.
// If we strip away the complex concurrency mechanics, it only does two things:
// 1. Forwards values: It reads a value from the upstream channel c and immediately pushes it into the returned channel valStream.
// 2. Exits safely: If the upstream channel c is closed or if the done channel cancels the operation, it stops everything, closes its output channel, and returns.
func orDone[T any](done <-chan struct{}, c <-chan T) <-chan T {
	valStream := make(chan T)

	go func() {
		defer close(valStream)

		for {
			// Read Select: Can I safely get the next piece of data without getting stuck?
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return // channel has been closed
				}
				// Write Select: Can I safely pass this data along without getting stuck?
				select {
				case valStream <- v: // forward v (blocks when no one reads from valStream)
				case <-done: // to prevent blocking (see line above)
					return
				}
			}
		}
	}()

	return valStream
}

func tee[T any](done <-chan struct{}, in <-chan T) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)

		// Leverages the orDone pattern to read from the in channel. This
		// guarantees that if the done channel is closed while waiting for input,
		// the loop breaks instantly and cleanly
		for val := range orDone(done, in) {
			// At the start of every new incoming value, fresh local copies of the
			// channels are created using shadowing. This is necessary because the
			// loop below is about to clear them to nil.
			out1, out2 := out1, out2

			for range 2 {
				select {
				case <-done:
					return
				case out1 <- val:
					// Dynamic Channel Disabling:
					// Sending to a nil channel in Go blocks indefinitely.
					// A send on a nil channel never becomes ready. So this marks the channel
					// as consumed after forwarding, loop will use the other one next.
					out1 = nil
				case out2 <- val:
					out2 = nil
				}
			}
		}
	}()

	return out1, out2
}

func main() {
	done := make(chan struct{})
	defer close(done)

	out1, out2 := tee(done, take(done, repeat(done, 1, 2), 4))

	for val1 := range out1 {
		fmt.Printf("out1: %v, out2: %v\n", val1, <-out2)
	}
}
