// Package main demonstrates golang.org/x/sync/errgroup, an Errgroup-style
// goroutine group that collects errors and (optionally) cancels a context on
// the first failure.
//
// errgroup.Group is like sync.WaitGroup, but with two additions:
//   1. It returns the first non-nil error from any goroutine.
//   2. errgroup.WithContext derives a context that is cancelled when the first
//      goroutine returns an error, so every other goroutine can observe
//      ctx.Done() and stop early.
//
// This contrasts with the steward/supervisor pattern in examples 4 and 5,
// where a failed child is *restarted* rather than cancelling the whole group.
// Errgroup is best for "all-or-nothing" workloads: fan-out, scatter-gather,
// map-reduce, or any operation where one failure means the whole thing aborts.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	exampleBasic()
	exampleWithContext()
	exampleFanIn()
	exampleSupervisorUnderErrgroup()
}

// exampleBasic shows the simplest errgroup usage: launch several goroutines,
// wait for all of them to finish, and collect the first error.
//
// Without a context, errgroup behaves like sync.WaitGroup + an error bucket.
func exampleBasic() {
	fmt.Println("\n=== Example 1: Basic errgroup (no context) ===")

	var g errgroup.Group

	for i := range 3 {
		g.Go(func() error {
			time.Sleep(time.Duration(i) * 100 * time.Millisecond)
			log.Printf("worker %d: done", i)
			return nil
		})
	}

	// Wait blocks until every goroutine started with g.Go has returned.
	if err := g.Wait(); err != nil {
		log.Printf("basic: finished with error: %v", err)
	} else {
		log.Println("basic: all workers succeeded")
	}
}

// exampleWithContext shows errgroup.WithContext.
//
// WithContext returns a Group and a derived context. If *any* goroutine
// returns a non-nil error, the context is cancelled. Every other goroutine
// can select on ctx.Done() and stop early.
//
// This is the key difference from sync.WaitGroup: errors propagate as
// cancellation, not just as a return value.
func exampleWithContext() {
	fmt.Println("\n=== Example 2: errgroup.WithContext ===")

	g, ctx := errgroup.WithContext(context.Background())

	// Worker A fails after 200 ms. This triggers context cancellation.
	g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		log.Println("worker A: failing now")
		return errors.New("worker A failed")
	})

	// Worker B keeps ticking until it observes ctx.Done().
	g.Go(func() error {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("worker B: still running")
			case <-ctx.Done():
				log.Printf("worker B: context cancelled (%v), stopping", ctx.Err())
				return ctx.Err() // returning ctx.Err() is conventional
			}
		}
	})

	if err := g.Wait(); err != nil {
		log.Printf("withcontext: group finished with error: %v", err)
	} else {
		log.Println("withcontext: all workers succeeded")
	}
}

// exampleFanIn shows a practical fan-in / scatter-gather pattern. Five workers
// run in parallel; worker 3 fails, which cancels the context. The watcher
// goroutine observes the cancellation and stops early instead of running to
// completion.
func exampleFanIn() {
	fmt.Println("\n=== Example 3: Fan-in with errgroup ===")

	g, ctx := errgroup.WithContext(context.Background())

	for i := range 5 {
		g.Go(func() error {
			time.Sleep(100 * time.Millisecond)
			if i == 3 {
				return fmt.Errorf("worker %d: simulated failure", i)
			}
			log.Printf("worker %d: succeeded", i)
			return nil
		})
	}

	// Watcher stops as soon as the context is cancelled by any worker error.
	g.Go(func() error {
		ticker := time.NewTicker(60 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("watcher: still waiting")
			case <-ctx.Done():
				log.Printf("watcher: context cancelled (%v)", ctx.Err())
				return ctx.Err()
			}
		}
	})

	if err := g.Wait(); err != nil {
		log.Printf("fanin: group finished with error: %v", err)
	}
}

// exampleSupervisorUnderErrgroup shows a real supervisor restart loop running
// inside errgroup.
//
// The supervisor runs a flaky worker. When the worker fails, the supervisor
// restarts it (up to maxRestarts times). If it fails too many times, the
// supervisor returns a fatal error. Either way, when the supervisor is done it
// cancels the context, which shuts down the sidecar goroutine.
//
// This combines the two patterns:
//   - supervisor: heals a worker by restarting it (examples 4 & 5)
//   - errgroup: cancels the whole group when healing is exhausted
func exampleSupervisorUnderErrgroup() {
	fmt.Println("\n=== Example 4: Supervisor under errgroup ===")

	// Use context.WithCancel directly so the supervisor can cancel ctx when it
	// finishes (success or exhaustion). errgroup.WithContext only cancels on
	// error, not on success.
	ctx, cancel := context.WithCancel(context.Background())
	g := new(errgroup.Group)

	const maxRestarts = 3

	// worker simulates a unit of work that fails on the first two attempts,
	// then succeeds on the third.
	worker := func(attempt int) error {
		time.Sleep(100 * time.Millisecond)
		if attempt < 3 {
			return fmt.Errorf("worker attempt %d: flaky failure", attempt)
		}
		log.Printf("worker attempt %d: succeeded", attempt)
		return nil
	}

	g.Go(func() error {
		defer cancel() // always cancel ctx when the supervisor is done

		restarts := 0
		attempt := 1

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err := worker(attempt); err != nil {
				log.Printf("supervisor: worker attempt %d failed: %v", attempt, err)

				if restarts >= maxRestarts {
					return fmt.Errorf("supervisor: exhausted %d restarts, giving up: %w", restarts, err)
				}

				restarts++
				attempt++
				log.Printf("supervisor: restarting worker (restart #%d)", restarts)
				continue
			}

			log.Printf("supervisor: worker healthy after %d restarts", restarts)
			return nil
		}
	})

	// A sidecar goroutine runs alongside and stops when the supervisor
	// cancels ctx (either on success or exhaustion).
	g.Go(func() error {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("sidecar: still running")
			case <-ctx.Done():
				log.Printf("sidecar: context cancelled (%v), stopping", ctx.Err())
				return nil // clean shutdown, not an error
			}
		}
	})

	if err := g.Wait(); err != nil {
		log.Printf("supervisor-under-errgroup: finished with error: %v", err)
	} else {
		log.Println("supervisor-under-errgroup: all done, no errors")
	}
}
