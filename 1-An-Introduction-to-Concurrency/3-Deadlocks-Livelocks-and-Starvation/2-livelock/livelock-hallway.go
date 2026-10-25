package main

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	cadence := sync.NewCond(&sync.Mutex{})

	// wake up all goroutines every 1ms (our metronome)
	go func() {
		for range time.Tick(1 * time.Millisecond) {
			cadence.Broadcast()
		}
	}()

	takeStep := func() {
		cadence.L.Lock()   // lock
		cadence.Wait()     // unlock, block until the metronome broadcasts, re-lock before returning
		cadence.L.Unlock() // unlock
	}

	// Only if Alice or Megan steps in one direction but the other person does not will this be a "Success" case.
	// If Alice and Megan walk in the same direction, there will be no "Success"; they block each other in the hallway.
	tryDir := func(dirName string, dir *int32, out *bytes.Buffer) bool {
		fmt.Fprintf(out, " %v", dirName)
		atomic.AddInt32(dir, 1)

		// Alice and Megan are waiting; they will be activated by our metronome at the same time (Broadcast).
		takeStep()

		// dir == 1: only this walker claimed this direction, so we succeed.
		// In the live-lock both claimed it, so dir == 2 and we fall through.
		if atomic.LoadInt32(dir) == 1 {
			fmt.Fprint(out, ". Success!")
			return true
		}

		// Alice and Megan are waiting; they will be activated by our metronome at the same time (Broadcast).
		takeStep()

		// Give up this walker's claim; when both back off, dir returns to 0.
		atomic.AddInt32(dir, -1)
		return false
	}

	// left and right are initially 0
	var left, right int32

	// try going left
	tryLeft := func(out *bytes.Buffer) bool { return tryDir("left", &left, out) }

	// try going right
	tryRight := func(out *bytes.Buffer) bool { return tryDir("right", &right, out) }

	// walk by trying to go left or right
	walk := func(walking *sync.WaitGroup, name string) {
		var out bytes.Buffer

		// Deferred calls run LIFO, so print before calling Done. Otherwise main
		// could return from Wait and exit before this walker's output is flushed.
		defer walking.Done()
		defer func() { fmt.Println(out.String()) }()

		fmt.Fprintf(&out, "%v is trying to scoot:", name)

		// first try left, then try right
		// Alice and Megan are trying the same steps at the same time. They block each other in the hallway.
		for range 3 {
			if tryLeft(&out) || tryRight(&out) {
				return
			}
		}
		fmt.Fprintf(&out, "\n%v tosses her hands up in exasperation!", name)
	}

	// Alice and Megan walking at the same time
	var peopleInHallway sync.WaitGroup
	peopleInHallway.Add(2)
	go walk(&peopleInHallway, "Alice")
	go walk(&peopleInHallway, "Megan")
	peopleInHallway.Wait()
}
