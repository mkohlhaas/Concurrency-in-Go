package main

import (
	"log"
	"os"
	"time"
)

type startGoroutineFn func(
	done <-chan struct{},
	pulseInterval time.Duration,
) (heartbeat <-chan struct{})

func newSteward(
	timeout time.Duration,
	s startGoroutineFn,
) startGoroutineFn {
	return func(
		done <-chan struct{},
		pulseInterval time.Duration,
	) (<-chan struct{}) {
		heartbeat := make(chan struct{})
		go func() {
			defer close(heartbeat)

			var wardDone chan struct{}
			var wardHeartbeat <-chan struct{}
			startWard := func() {
				wardDone = make(chan struct{})
				wardHeartbeat = s(wardDone, timeout/2)
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
					case <-timeoutSignal:
						log.Println("steward: ward unhealthy; restarting")
						if wardDone != nil {
							close(wardDone)
						}
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

		return heartbeat
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	doWorkFn := func(
		done <-chan struct{},
		_ time.Duration,
	) (<-chan struct{}) {
		log.Println("ward: restarting")
		heartbeat := make(chan struct{})
		go func() {
			defer close(heartbeat)

			localDone := make(chan struct{})
			time.AfterFunc(13*time.Second, func() {
				log.Println("ward: halting by steward request")
				close(localDone)
			})

			for i := 0; i < 10; i++ {
				select {
				case heartbeat <- struct{}{}:
				case <-localDone:
					return
				case <-done:
					return
				}
				time.Sleep(1 * time.Second)
			}
		}()

		return heartbeat
	}

	doWorkWithSteward := newSteward(4*time.Second, doWorkFn)

	done := make(chan struct{})
	time.AfterFunc(9*time.Second, func() {
		log.Println("main: halting steward and ward")
		close(done)
	})

	for range doWorkWithSteward(done, 4*time.Second) {
		log.Println("main: received heartbeat")
	}
}
