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
			time.AfterFunc(13*time.Second, func() {
				log.Println("ward: halting by steward request")
				close(localDone)
			})

			for range 10 {
				select {
				case heartbeat <- struct{}{}:
				case <-localDone:
					return
				case <-done:
					return
				}
				time.Sleep(1 * time.Second)
			}

			for i := range 10 {
				select {
				case results <- fmt.Errorf("error from iteration %v", i):
				case <-localDone:
					return
				case <-done:
					return
				}
				time.Sleep(1 * time.Second)
			}
		}()

		return heartbeat, results
	}

	doWorkWithSteward := newSteward(4*time.Second, doWorkFn)

	done := make(chan struct{})
	time.AfterFunc(9*time.Second, func() {
		log.Println("main: halting steward and ward")
		close(done)
	})

	heartbeatStream, resultsStream := doWorkWithSteward(done, 4*time.Second)
	for {
		select {
		case _, ok := <-heartbeatStream:
			if !ok {
				return
			}
			log.Println("main: received heartbeat")
		case err, ok := <-resultsStream:
			if !ok {
				return
			}
			if err != nil {
				log.Printf("main: received error: %v\n", err)
			}
		}
	}
}
