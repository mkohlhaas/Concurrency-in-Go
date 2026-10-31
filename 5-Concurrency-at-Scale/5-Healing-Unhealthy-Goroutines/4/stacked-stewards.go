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
// still alive. Only the goroutine that owns the work may send on it.
type heartbeat <-chan struct{}

// doneChan is a receive-only stop signal. The owner closes it to broadcast
// "stop"; consumers only ever select on it.
type doneChan <-chan struct{}

// startGoroutineFn is the contract for anything a steward can watch over. That
// is usually a ward, but because newSteward returns a startGoroutineFn too, it
// can also be another steward.
type startGoroutineFn func(done doneChan, pulseInterval time.Duration) heartbeat

// logGoroutines logs who is doing what, plus a snapshot of how many goroutines
// exist at that moment. The count is a snapshot, and goroutines that are
// blocked forever are still counted, which is the point of this example.
func logGoroutines(who, format string, args ...any) {
	log.Printf("%s: %s (goroutines = %d)", who, fmt.Sprintf(format, args...), runtime.NumGoroutine())
}

// newSteward wraps start in a named supervisor. It starts the ward, watches the
// ward's heartbeat channel, and replaces the ward whenever the ward stops
// proving it is alive. The returned function has the same signature as the
// thing it wraps, so a steward is itself monitorable: stewards can be stacked.
func newSteward(name string, timeout time.Duration, start startGoroutineFn) startGoroutineFn {
	return newStewardWithWedge(name, timeout, start, nil)
}

// newStewardWithWedge is newSteward with a fault hook used only by this demo.
// After every heartbeat the steward sends upstream, wedge(pulses) is called;
// returning true wedges the steward itself, meaning its monitor loop blocks
// forever. A wedged steward stops pulsing and stops watching its ward, but it
// never returns, so it never closes its heartbeat channel either.
func newStewardWithWedge(
	name string,
	timeout time.Duration,
	start startGoroutineFn,
	wedge func(pulses int) bool,
) startGoroutineFn {
	// Generation numbering is only for readable logs: a factory that is called
	// again (because an outer steward restarted it) produces "inner #2", etc.
	var generation atomic.Int64

	return func(done doneChan, pulseInterval time.Duration) heartbeat {
		name := fmt.Sprintf("%s #%d", name, generation.Add(1))
		hb := make(chan struct{})

		go func() {
			defer close(hb)

			// The steward keeps its own done channel for the ward, so that it
			// can kill and replace the ward without shutting itself down.
			var (
				wardDone      chan struct{}
				wardHeartbeat heartbeat
			)
			startWard := func() {
				wardDone = make(chan struct{})
				// Give the ward half the timeout, so it gets two chances to
				// report in before the steward gives up on it.
				wardHeartbeat = start(wardDone, timeout/2)
			}
			killWard := func() { close(wardDone) }
			// awaitWardStop blocks until the ward has actually stopped or the
			// grace period expires, so a wedged ward cannot hold the steward
			// hostage forever. The same bound is what lets an outer steward
			// survive a wedged inner steward.
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
			log.Printf("%s: started; watching ward", name)

			pulse := time.NewTicker(pulseInterval)
			defer pulse.Stop()

			pulses := 0
			for {
				// One timeout per ward generation: the clock starts when the
				// ward starts (or reports in) and is only reset by the ward.
				timeoutSignal := time.After(timeout)

			wardCheckin:
				for {
					select {
					case <-pulse.C:
						// Report our own health upstream, never blocking.
						select {
						case hb <- struct{}{}:
						default:
						}
						pulses++
						if wedge != nil && wedge(pulses) {
							logGoroutines(name, "wedged after %d heartbeats; no longer pulsing", pulses)
							select {} // stuck, not exited: the fault an outer steward heals
						}
					case _, ok := <-wardHeartbeat:
						if !ok {
							// The ward returned and closed its heartbeat.
							log.Printf("%s: ward exited; restarting", name)
							killWard()
							awaitWardStop()
							startWard()
							log.Printf("%s: ward replaced", name)
							break wardCheckin
						}
						// The ward is alive: reset the timeout.
						break wardCheckin
					case <-timeoutSignal:
						logGoroutines(name, "ward unhealthy; restarting")
						killWard()
						awaitWardStop()
						startWard()
						log.Printf("%s: ward replaced", name)
						break wardCheckin
					case <-done:
						killWard()
						awaitWardStop()
						log.Printf("%s: ward stopped; returning", name)
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

	const (
		innerTimeout = 500 * time.Millisecond  // silence the inner steward tolerates
		outerTimeout = 1500 * time.Millisecond // silence the outer steward tolerates
		mainPulse    = 500 * time.Millisecond  // how often the outer steward reports to main
	)

	logGoroutines("main", "starting")

	// The ward does the actual work. Here it just reports in on schedule.
	var wardID atomic.Int64
	doWorkFn := func(done doneChan, pulseInterval time.Duration) heartbeat {
		id := wardID.Add(1)
		log.Printf("ward #%d: starting; reporting in every %v", id, pulseInterval)

		hb := make(chan struct{})
		go func() {
			defer close(hb)
			logGoroutines(fmt.Sprintf("ward #%d", id), "running")

			report := time.NewTicker(pulseInterval)
			defer report.Stop()

			for pulses := 1; ; pulses++ {
				// Report in, but always stay interruptible by done.
				select {
				case hb <- struct{}{}:
					log.Printf("ward #%d: heartbeat %d", id, pulses)
				case <-done:
					log.Printf("ward #%d: halting", id)
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

	// Inject exactly one fault: the first inner steward wedges after three
	// heartbeats. The flag makes sure only the first generation is affected, so
	// the replacement the outer steward starts can run healthy.
	var wedged atomic.Bool
	wedge := func(pulses int) bool {
		return pulses >= 3 && wedged.CompareAndSwap(false, true)
	}

	// The stack: outer watches inner, inner watches the ward.
	inner := newStewardWithWedge("inner", innerTimeout, doWorkFn, wedge)
	outer := newSteward("outer", outerTimeout, inner)

	done := make(chan struct{})
	time.AfterFunc(10*time.Second, func() {
		log.Println("main: halting outer steward")
		close(done)
	})

	hb := outer(done, mainPulse)
	log.Println("main: outer steward started")

	// The outer steward keeps pulsing even while it heals the inner one, so
	// main never sees the outage.
	for range hb {
		log.Println("main: heartbeat from outer steward")
	}

	// The wedged inner steward and its blocked ward are still counted here:
	// Go cannot kill a goroutine, so only process exit reclaims them.
	logGoroutines("main", "done")
}
