package main

import (
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

	type startGoroutineFn func(done <-chan any, pulseInterval time.Duration) (heartbeat <-chan any)

	// The steward itself returns a startGoroutineFn indicating that the steward itself is also monitorable!!!
	newSteward := func(timeout time.Duration, startGoroutine startGoroutineFn) startGoroutineFn {
		return func(done <-chan any, pulseInterval time.Duration) <-chan any {
			heartbeat := make(chan any)
			// start steward
			go func() {
				defer close(heartbeat)

				var wardDone chan any
				var wardHeartbeat <-chan any

				startWard := func() {
					// use another done channel on top of that provided by the ward
					// necessary because you cannot close receive-only channel done
					wardDone = make(chan any) // read-write channel
					wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
				}

				// start ward in its own process (see doWorkWithSteward)
				startWard()
				pulse := time.Tick(pulseInterval)

			monitorLoop:
				for {
					timeoutSignal := time.After(timeout)
					for {
						select {
						case <-pulse: // steward sends heartbeat
							select {
							case heartbeat <- struct{}{}:
							default:
							}
						case <-wardHeartbeat: // receiving heartbeat from ward (our ward doesn't have a heartbeat!!!)
							continue monitorLoop // refresh timeoutSignal in the new iteration
						case <-timeoutSignal:
							log.Println("steward: ward unhealthy; restarting")
							close(wardDone) // shutdown ward
							startWard()     // restart ward; creates new ward channels
							continue monitorLoop
						case <-done: // ward has been shutdown -> shutdown steward
							return
						}
					}
				}
			}()

			return heartbeat
		}
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	// this ward does not send any heartbeats
	// also no timeouts used
	doWork := func(done <-chan any, _timeout time.Duration) <-chan any {
		log.Println("new ward: Hello, I'm the irresponsible ward!")
		go func() {
			<-done
			log.Println("old ward: I am halting.")
		}()
		return nil // returns a nil heartbeat channel
	}

	doWorkWithSteward := newSteward(4*time.Second, doWork) // 4 sec. = time given the ward to send heartbeats

	done := make(chan any)
	time.AfterFunc(1*time.Minute, func() { // shutting everything down after a minute
		log.Println("main: halting ward and steward.")
		close(done)
	})

	// check heartbeat of steward
	for range doWorkWithSteward(done, 1*time.Second) { // 1 sec. = heartbeat interval of steward
		log.Println("main: got heartbeat from steward.")
	}

	log.Println("Done")
}
