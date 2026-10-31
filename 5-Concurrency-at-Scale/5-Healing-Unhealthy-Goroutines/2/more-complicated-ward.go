package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// returns heartbeat channel and results channel
type startGoroutineFn func(
	done <-chan struct{},
	pulseInterval time.Duration,
) (heartbeat <-chan struct{}, results <-chan error)

// ------- //
// Steward //
// ------- //

func newSteward(
	timeout time.Duration,
	s startGoroutineFn,
) startGoroutineFn {
	return func(
		done <-chan struct{},
		pulseInterval time.Duration,
	) (<-chan struct{}, <-chan error) {
		heartbeat := make(chan struct{})
		results := make(chan error)

		go func() {
			defer close(heartbeat)
			defer close(results)

			var wardDone chan struct{}
			var wardHeartbeat <-chan struct{}
			var wardResults <-chan error
			startWard := func() {
				log.Println("steward: starting ward")
				wardDone = make(chan struct{})
				wardHeartbeat, wardResults = s(wardDone, timeout/2)
			}
			startWard()
			pulse := time.Tick(pulseInterval)

		monitorLoop:
			for {
				timeoutSignal := time.After(timeout)

				for {
					select {
					case <-pulse:
						select {
						case heartbeat <- struct{}{}:
						default:
						}
					case <-wardHeartbeat:
						continue monitorLoop
					case err := <-wardResults:
						results <- err
						continue monitorLoop
					case <-timeoutSignal:
						log.Println("steward: ward unhealthy; restarting")
						close(wardDone)
						startWard()
						continue monitorLoop
					case <-done:
						log.Println("steward: received done, halting ward")
						if wardDone != nil {
							close(wardDone)
						}
						return
					}
				}
			}
		}()

		return heartbeat, results
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	// ---- //
	// Ward //
	// ---- //

	doWorkFn := func(
		done <-chan struct{},
		_ time.Duration,
	) (<-chan struct{}, <-chan error) {
		log.Println("ward: restarting")
		heartbeat := make(chan struct{})
		results := make(chan error)
		go func() {
			defer close(heartbeat)
			defer close(results)

			localDone := make(chan struct{})
			time.AfterFunc(20*time.Second, func() {
				log.Println("ward: halting by steward request (localDone)")
				close(localDone)
			})

			for range 10 {
				log.Println("ward: about to send heartbeat")
				select {
				case heartbeat <- struct{}{}:
					log.Println("ward: sent heartbeat")
				case <-localDone:
					log.Println("ward: received localDone, halting")
					return
				case <-done:
					log.Println("ward: received done, halting")
					return
				}
				select {
				case <-time.After(1 * time.Second):
				case <-localDone:
					log.Println("ward: received localDone during sleep, halting")
					return
				case <-done:
					log.Println("ward: received done during sleep, halting")
					return
				}
			}

			for i := range 10 {
				log.Println("ward: about to send error")
				select {
				case results <- fmt.Errorf("error from iteration %v", i):
					log.Println("ward: sent error")
				case <-localDone:
					log.Println("ward: received localDone, halting")
					return
				case <-done:
					log.Println("ward: received done, halting")
					return
				}
				select {
				case <-time.After(1 * time.Second):
				case <-localDone:
					log.Println("ward: received localDone during sleep, halting")
					return
				case <-done:
					log.Println("ward: received done during sleep, halting")
					return
				}
			}
		}()

		return heartbeat, results
	}

	doWorkWithSteward := newSteward(25*time.Second, doWorkFn)

	done := make(chan struct{})
	time.AfterFunc(3*time.Second, func() {
		log.Println("main: halting steward and ward")
		close(done)
	})

	heartbeatStream, resultsStream := doWorkWithSteward(done, 4*time.Second)
	stop := false
	for !stop {
		select {
		case _, ok := <-heartbeatStream:
			if !ok {
				stop = true
			} else {
				log.Println("main: received heartbeat")
			}
		case err, ok := <-resultsStream:
			if !ok {
				stop = true
			} else if err != nil {
				log.Printf("main: received error: %v\n", err)
			}
		}
	}
	// Give the ward a moment to react to done before main exits.
	time.Sleep(2 * time.Second)
}
