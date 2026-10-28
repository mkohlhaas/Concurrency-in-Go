package main

import (
	"fmt"
)

// main demonstrates ad-hoc confinement.
//
// Ad-hoc confinement is the weakest form of confinement: there is no type or
// language mechanism stopping other goroutines from touching shared data.
// Instead, the code relies on a convention that only one goroutine accesses
// the data at a time. Here, `data` is written only by main (once, before the
// goroutine starts) and read only by the loopData goroutine, so no mutex or
// other synchronization is needed to protect it.
func main() {
	// The slice is created and initialized here; after this point no other
	// goroutine writes to it, which is what makes the sharing safe.
	data := make([]int, 4)

	// loopData sends each element of data on the channel it receives.
	// It only reads data, and it is the sole goroutine reading it, so the
	// access is ad-hoc confined.
	loopData := func(handleData chan<- int) {
		// Closing the channel signals the consumer (the range loop below)
		// that no more values are coming.
		defer close(handleData)

		for i := range data {
			handleData <- data[i]
		}
	}

	// Unbuffered channel: each send here synchronizes with a receive in the
	// consumer loop, so the hand-off is safe even though the data itself is
	// only confined by convention.
	handleData := make(chan int)
	go loopData(handleData)

	// range over the channel receives all values until loopData closes it.
	for num := range handleData {
		fmt.Println(num)
	}
}
