package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	// stdoutBuff collects all output so that it can be written in one go.
	// Writing to a buffer avoids interleaved output between the producer
	// goroutine and main.
	var stdoutBuff bytes.Buffer

	// Deferred functions run in LIFO order when main returns, so the buffer
	// is flushed to stdout as the very last step. Writing directly to
	// os.Stdout inside the goroutine could interleave lines, whereas the
	// buffer is only read here after everything else is done.
	defer func() { _, _ = stdoutBuff.WriteTo(os.Stdout) }()

	// A buffered channel with capacity 4. Sends block only once the buffer is
	// full, so the producer can run ahead of the consumer by up to four
	// values before it has to wait.
	intStream := make(chan int, 4)

	go func() {
		// Defers run LIFO: "Producer Done." is printed first, then the
		// channel is closed. Closing signals the consumer's range loop to
		// stop once the buffered values have been drained.
		defer close(intStream)
		defer fmt.Fprintln(&stdoutBuff, "Producer Done.")

		// Send five values. The first four fit in the buffer without
		// blocking; the fifth blocks until the consumer receives one.
		for i := range 20 {
			fmt.Fprintf(&stdoutBuff, "Sending: %d\n", i)
			intStream <- i
		}
	}()

	// Range receives values until the channel is closed and empty. Because
	// the buffer decouples the two sides, some "Sending" lines may appear
	// before the matching "Received" line and vice versa.
	for i := range intStream {
		fmt.Fprintf(&stdoutBuff, "Received: %v\n", i)
	}
}
