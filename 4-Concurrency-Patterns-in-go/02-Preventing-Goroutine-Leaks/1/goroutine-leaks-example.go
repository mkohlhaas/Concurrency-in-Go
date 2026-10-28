package main

import "fmt"

func main() {
	// doWork takes a read-only channel of strings and returns a channel that
	// signals completion. The caller can use the returned channel to know when
	// the internal goroutine has finished.
	doWork := func(strings <-chan string) <-chan struct{} {
		// done is closed when the goroutine below exits, so anyone reading
		// from it unblocks at that point.
		done := make(chan struct{})

		// Spawn the worker goroutine. Deferred calls run in LIFO order, so
		// close(done) runs first, then the log message is printed.
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(done)

			// range over a channel blocks until a value is received or the
			// channel is closed. This is the source of the leak below.
			for s := range strings {
				// Do something interesting
				fmt.Println(s)
			}
		}()

		return done
	}

	// 1. Use this...
	//
	// Passing a nil channel means the range above blocks forever, since
	// receiving from a nil channel never returns. The goroutine can never
	// reach its deferred calls, so it never exits: a goroutine leak.
	{
		doWork(nil)
	}

	// 2. ... or this.
	//
	// Normally main would wait for the goroutine to finish by receiving from
	// the returned channel.
	//
	// That would block forever instead of completing, because done is
	// never closed. Seeing that hang is one way to observe the leak.
	//
	// {
	// 	done := doWork(nil)
	// 	<-done
	// }

	// Perhaps more work is done here
	fmt.Println("Done.")
}
