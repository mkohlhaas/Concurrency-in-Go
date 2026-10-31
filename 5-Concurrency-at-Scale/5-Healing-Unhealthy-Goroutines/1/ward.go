package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

// heartbeat is the receive side of the channel a goroutine uses to prove it is
// still alive. It is deliberately receive-only: only the goroutine that owns
// the work may send on it, and everybody else just watches. The owner keeps the
// bidirectional channel locally and returns it as a heartbeat.
type heartbeat <-chan struct{}

// doneChan is a receive-only stop signal. The owner closes it to broadcast
// "stop"; consumers only ever select on it.
type doneChan <-chan struct{}

// startGoroutineFn is the contract for anything the steward can watch over:
// it is handed a done channel to stop on and the interval at which it is
// expected to report in, and it returns the heartbeat channel it reports on.
//
// Both channels are receive-only struct{} channels, but they carry different
// signals: done is closed to broadcast "stop", while heartbeat is sent on once
// per pulse to say "alive".
//
// A goroutine that follows this contract can be wrapped in newSteward without
// the steward knowing anything about what the work actually is.
type startGoroutineFn func(
	done doneChan,
	pulseInterval time.Duration,
) heartbeat

// logGoroutines logs who is doing what, plus a snapshot of how many goroutines
// exist at that moment, the caller included.
//
// The count is only a snapshot: a goroutine that has returned but has not been
// reaped yet is still counted, so a restart boundary can briefly show one more
// goroutine than the steady state.
func logGoroutines(who, format string, args ...any) {
	log.Printf("%s: %s (goroutines = %d)", who, fmt.Sprintf(format, args...), runtime.NumGoroutine())
}

// newSteward wraps start in a supervisor. It starts the ward, watches the ward's
// heartbeat channel, and replaces the ward whenever the ward stops proving it
// is alive. The returned function has the same signature as the thing it wraps,
// so a steward is itself monitorable: stewards can be stacked.
func newSteward(
	timeout time.Duration,
	start startGoroutineFn,
) startGoroutineFn {
	return func(
		done doneChan,
		pulseInterval time.Duration,
	) heartbeat {
		// The steward keeps the send side locally; callers only ever see the
		// receive-only heartbeat type.
		hb := make(chan struct{})

		go func() {
			defer close(hb)

			// The steward keeps its own done channel for the ward, so that it can
			// kill and replace the ward without shutting itself down.
			var (
				wardDone      chan struct{}
				wardHeartbeat heartbeat
			)
			startWard := func() {
				wardDone = make(chan struct{})
				// Give the ward half the timeout, so it gets two chances to report
				// in before the steward gives up on it.
				wardHeartbeat = start(wardDone, timeout/2)
			}
			killWard := func() { close(wardDone) }
			// awaitWardStop blocks until the ward has actually stopped (a well-behaved
			// ward closes its heartbeat channel on the way out) or the grace period
			// expires, so that a slow ward cannot hold the steward hostage.
			awaitWardStop := func() {
				grace := time.After(timeout)
				for {
					select {
					case _, ok := <-wardHeartbeat:
						if !ok {
							return
						}
					case <-grace:
						return
					}
				}
			}

			startWard()
			logGoroutines("steward", "started; watching ward")

			pulse := time.NewTicker(pulseInterval)
			defer pulse.Stop()

			for {
				// One timeout per ward generation: the clock starts when the ward
				// starts (or reports in) and is only reset by the ward itself.
				timeoutSignal := time.After(timeout)

			wardCheckin:
				for {
					select {
					case <-pulse.C:
						// Report our own health upstream. Never block on it: whoever
						// is watching us may have stopped listening.
						select {
						case hb <- struct{}{}:
						default:
						}
					case _, ok := <-wardHeartbeat:
						if !ok {
							// A closed heartbeat channel means the ward returned.
							// There is nothing to wait for, so replace it at once.
							log.Println("steward: ward exited; restarting")
							killWard()
							awaitWardStop()
							startWard()
							logGoroutines("steward", "ward replaced")
							break wardCheckin
						}
						// The ward is alive: reset the timeout and keep waiting.
						break wardCheckin
					case <-timeoutSignal:
						// Silence for a whole timeout: the ward is deadlocked or
						// otherwise stuck. Kill it and start a fresh one.
						logGoroutines("steward", "ward unhealthy; restarting")
						killWard()
						awaitWardStop()
						startWard()
						logGoroutines("steward", "ward replaced")
						break wardCheckin
					case <-done:
						killWard()
						awaitWardStop()
						logGoroutines("steward", "ward stopped; returning")
						return
					}
				}
			}
		}()
		return hb
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	logGoroutines("main", "starting")

	const (
		timeout       = 2 * time.Second // how long a ward may stay silent
		pulseInterval = 2 * time.Second // how often the steward reports to us
	)

	// doWorkFn is the ward. It reports in every pulseInterval, as the steward
	// asked it to. To make the pattern visible, the first two wards are scripted
	// to fail in the two ways a ward can fail:
	//
	//	ward #1: deadlocks - it stops reporting in but never returns
	//	ward #2: crashes   - it returns, closing its heartbeat channel
	//	ward #3: is healthy until we shut everything down
	//
	// A real ward would be the same code every time; the fault is injected here
	// only so the steward has something to heal.
	var wardID atomic.Int64

	doWorkFn := func(
		done doneChan,
		pulseInterval time.Duration,
	) heartbeat {
		id := wardID.Add(1)
		log.Printf("ward #%d: starting; reporting in every %v", id, pulseInterval)

		hb := make(chan struct{})
		go func() {
			defer close(hb)

			report := time.NewTicker(pulseInterval)
			defer report.Stop()

			for pulses := 1; ; pulses++ {
				// Report in. Select on done as well, so the steward can always
				// interrupt us, even if nobody is reading heartbeats right now.
				select {
				case hb <- struct{}{}:
					log.Printf("ward #%d: heartbeat %d", id, pulses)
				case <-done:
					log.Printf("ward #%d: halting", id)
					return
				}

				// Inject this ward's failure, if it has one.
				switch {
				case id == 1 && pulses == 3:
					log.Printf("ward #%d: deadlocked; no longer reporting in", id)
					<-done // only the steward can get us out of this
					log.Printf("ward #%d: halting", id)
					return
				case id == 2 && pulses == 3:
					log.Printf("ward #%d: crashed", id)
					return
				}

				select {
				case <-report.C:
				case <-done:
					log.Printf("ward #%d: halting", id)
					return
				}
			}
		}()

		return hb
	}

	// Wrap the ward in a steward. The steward is a startGoroutineFn too, so from
	// here on it is just another monitored goroutine.
	doWorkWithSteward := newSteward(timeout, doWorkFn)

	// Give the demo long enough to show both failures and a healthy recovery.
	done := make(chan struct{})
	time.AfterFunc(15*time.Second, func() {
		// This callback runs in its own goroutine, so the count below is one
		// higher than the steady state: main + steward + ward + the timer.
		logGoroutines("main", "halting steward and ward")
		close(done)
	})

	// The steward closes our channel when it stops, which ends this loop.
	heartbeat := doWorkWithSteward(done, pulseInterval)
	// The steward starts the ward on its own goroutine, so this count can read 2
	// (ward not started yet) or 3. The steward logs again right after startWard,
	// where the ward goroutine is guaranteed to exist.
	logGoroutines("main", "steward started")

	for range heartbeat {
		log.Println("main: received heartbeat from steward")
	}

	logGoroutines("main", "done")
}
