package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	var or func(channels ...<-chan any) <-chan any
	or = func(channels ...<-chan any) <-chan any {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}

		orDone := make(chan any)
		go func() {
			defer close(orDone)

			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()
		return orDone
	}

	orDone := func(done, c <-chan any) <-chan any {
		valStream := make(chan any)
		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if ok == false {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	type startGoroutineFn func(done <-chan any, pulseInterval time.Duration) (heartbeat <-chan any)

	newSteward := func(timeout time.Duration, startGoroutine startGoroutineFn) startGoroutineFn {
		return func(done <-chan any, pulseInterval time.Duration) <-chan any {
			heartbeat := make(chan any)
			go func() {
				defer close(heartbeat)

				var wardDone chan any
				var wardHeartbeat <-chan any
				startWard := func() {
					wardDone = make(chan any)
					wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
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
							close(wardDone)
							startWard()
							continue monitorLoop
						case <-done:
							return
						}
					}
				}
			}()

			return heartbeat
		}
	}

	take := func(done <-chan any, valueStream <-chan any, num int) <-chan any {
		takeStream := make(chan any)
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}
		}()
		return takeStream
	}

	bridge := func(done <-chan any, chanStream <-chan <-chan any) <-chan any {
		valStream := make(chan any)
		go func() {
			defer close(valStream)
			for {
				var stream <-chan any
				select {
				case maybeStream, ok := <-chanStream:
					if ok == false {
						return
					}
					stream = maybeStream
				case <-done:
					return
				}
				for val := range orDone(done, stream) {
					select {
					case valStream <- val:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	doWorkFn := func(done <-chan any, intList ...int) (startGoroutineFn, <-chan any) {
		intChanStream := make(chan (<-chan any))
		intStream := bridge(done, intChanStream)
		doWork := func(done <-chan any, pulseInterval time.Duration) <-chan any {
			intStream := make(chan any)
			heartbeat := make(chan any)
			go func() {
				defer close(intStream)
				select {
				case intChanStream <- intStream: // When (re-)starting wards we just insert the new intStream into the bridge!!! client just keeps on using intChanStream!!!
				case <-done:
					return
				}

				pulse := time.Tick(pulseInterval)

				for {
				valueLoop:
					for _, intVal := range intList {
						if intVal < 0 { // on a negative value stop the ward simulating an unhealthy ward
							log.Printf("negative value: %v\n", intVal)
							return
						}

						for {
							select {
							case <-pulse:
								select {
								case heartbeat <- struct{}{}: // this ward sends a heartbeat
								default:
								}
							case intStream <- intVal:
								continue valueLoop
							case <-done:
								return
							}
						}
					}
				}
			}()
			return heartbeat
		}
		return doWork, intStream
	}

	log.SetFlags(log.Ltime | log.LUTC)
	log.SetOutput(os.Stdout)

	done := make(chan any)
	defer close(done)

	doWork, intStream := doWorkFn(done, 1, 2, 1, -3, 4, 5)
	doWorkWithSteward := newSteward(1*time.Second, doWork)
	doWorkWithSteward(done, 1*time.Hour)

	for intVal := range take(done, intStream, 25) {
		fmt.Printf("Received: %v\n", intVal)
	}
}
