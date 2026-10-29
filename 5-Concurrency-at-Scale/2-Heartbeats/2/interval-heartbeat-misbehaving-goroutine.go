package main

import (
	"fmt"
	"time"
)

func main() {
	doWork := func(done <-chan struct{}, pulseInterval time.Duration) (<-chan struct{}, <-chan time.Time) {
		heartbeat := make(chan struct{})
		results := make(chan time.Time)

		go func() {
			pulse := time.NewTicker(pulseInterval)
			defer pulse.Stop()
			workGen := time.NewTicker(2 * pulseInterval)
			defer workGen.Stop()

			sendPulse := func() {
				select {
				case heartbeat <- struct{}{}:
				default:
				}
			}

			sendResult := func(r time.Time) {
				for {
					select {
					case <-done:
						return
					// keep sending heartbeats while blocked delivering a result,
					// so the liveness watchdog doesn't trip during result delivery
					case <-pulse.C:
						sendPulse()
					case results <- r:
						return
					}
				}
			}

			// will stop after 10 iterations
			for range 10 {
				select {
				case <-done:
					return
				case <-pulse.C:
					sendPulse()
				case r := <-workGen.C:
					sendResult(r)
				}
			}
		}()

		return heartbeat, results
	}

	// run the program for this many seconds before shutting down
	const runForSeconds = 5
	done := make(chan struct{})
	time.AfterFunc(runForSeconds*time.Second, func() { close(done) })

	// livenessTimeout is the max silence we tolerate before declaring the
	// worker dead. pulseInterval = livenessTimeout/2 so a heartbeat arrives
	// every 250 ms, well within the 500 ms watchdog.
	const livenessTimeout = 500 * time.Millisecond
	heartbeat, results := doWork(done, livenessTimeout/2)

	for {
		select {
		// liveness watchdog: resets every loop iteration because a heartbeat
		// arrives every 250 ms (livenessTimeout/2), so this case only wins if
		// no heartbeat or result arrives for a full 500 ms — i.e. the worker is
		// stuck or dead. It is NOT an "exit after 500 ms" deadline.
		// It's recreated on every loop iteration!
		case <-time.After(livenessTimeout):
			fmt.Println("worker goroutine is not healthy!")
			return
		// read heartbeat
		case _, ok := <-heartbeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		// read result
		case r, ok := <-results:
			if !ok {
				return
			}
			fmt.Printf("results %v\n", r.Second())
		}
	}
}
