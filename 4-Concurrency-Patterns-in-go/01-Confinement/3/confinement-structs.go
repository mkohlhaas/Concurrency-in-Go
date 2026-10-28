package main

import (
	"bytes"
	"fmt"
	"sync"
)

// This example demonstrates confinement: keeping mutable data local so that
// multiple goroutines can work on it without sharing it through a lock or
// channel.
//
// The slice data is only referenced inside main. Each goroutine receives a
// sub-slice of the same backing array, but the two ranges never overlap and
// are only read, so there is no data race. Confinement here is enforced by
// convention (which index range each goroutine gets), not by the compiler.
func main() {
	// printData writes its given bytes into a buffer owned by this invocation
	// and then prints them. Because buff is local to the call, each goroutine
	// gets its own buffer and no synchronization is needed for it.
	printData := func(wg *sync.WaitGroup, data []byte) {
		defer wg.Done()

		var buff bytes.Buffer

		for _, b := range data {
			fmt.Fprintf(&buff, "%c", b)
		}
		fmt.Println(buff.String())
	}

	var wg sync.WaitGroup
	// We launch two goroutines, so expect two Done calls.
	wg.Add(2)

	// data lives on the main goroutine's stack/frame; the goroutines only
	// borrow read-only views of it.
	data := []byte("golang")

	// Split the slice into two disjoint, non-overlapping regions:
	// data[:3] is "gol" and data[3:] is "ang".
	go printData(&wg, data[:3])
	go printData(&wg, data[3:])

	// Wait for both goroutines to finish printing before main returns.
	wg.Wait()
}
