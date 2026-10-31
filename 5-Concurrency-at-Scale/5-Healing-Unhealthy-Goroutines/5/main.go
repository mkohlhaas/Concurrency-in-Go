// Package main demonstrates an Erlang-style supervision tree in Go.
//
// A Supervisor watches over multiple children (workers or sub-supervisors)
// and restarts them when they stop proving they are alive. Because a
// Supervisor implements startGoroutineFn, supervisors can nest: a supervisor
// can itself be a child of another supervisor. That is what makes a
// supervision tree.
//
// Restart strategies mirror Erlang's:
//   - one_for_one: restart only the child that failed.
//   - one_for_all: restart every child if any one fails.
//   - rest_for_all: restart the failed child and every child started after it.
//
// Go cannot kill a goroutine, so a wedged worker (one that ignores cancellation)
// is detected by silence: its heartbeat channel stops pulsing. The supervisor
// restarts it by cancelling its context and starting a fresh goroutine. The
// wedged goroutine leaks unless it cooperates with context cancellation; only
// process-level reclamation can force it out.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// heartbeat is the receive side of the channel a supervised goroutine uses to
// prove it is still alive. Only the goroutine itself may send on it.
type heartbeat <-chan struct{}

// startGoroutineFn starts supervised work and returns the work's heartbeat.
// A Supervisor implements this signature, so a supervisor can be a child of
// another supervisor — that is what makes a supervision tree.
type startGoroutineFn func(ctx context.Context, pulseInterval time.Duration) heartbeat

// restartStrategy mirrors Erlang's supervisor restart strategies.
type restartStrategy int

const (
	// OneForOne restarts only the child that failed.
	OneForOne restartStrategy = iota
	// OneForAll restarts every child if any one fails.
	OneForAll
	// RestForAll restarts the failed child and every child started after it.
	RestForAll
)

func (r restartStrategy) String() string {
	switch r {
	case OneForOne:
		return "one_for_one"
	case OneForAll:
		return "one_for_all"
	default:
		return "rest_for_all"
	}
}

func logGoroutines(who, format string, args ...any) {
	log.Printf("%s: %s (goroutines = %d)", who, fmt.Sprintf(format, args...), runtime.NumGoroutine())
}

// child is a single supervised entity: either a worker or a sub-supervisor.
type child struct {
	name      string
	start     startGoroutineFn
	interval  time.Duration
	hb        heartbeat
	cancel    context.CancelFunc
	restarts  atomic.Int32
	lastPulse atomic.Int64 // unix nano of the most recent heartbeat
}

// Supervisor watches several children and restarts them when they fall silent.
type Supervisor struct {
	name     string
	timeout  time.Duration // silence beyond this => child is unhealthy
	strategy restartStrategy
	children []*child

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.Mutex
}

// NewSupervisor creates a supervisor with the given timeout and strategy.
// The supervisor does not start until Run is called with a parent context.
func NewSupervisor(name string, timeout time.Duration, strategy restartStrategy) *Supervisor {
	return &Supervisor{
		name:     name,
		timeout:  timeout,
		strategy: strategy,
	}
}

// Add registers a child to supervise. interval is how often the child pulses.
func (s *Supervisor) Add(name string, start startGoroutineFn, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.children = append(s.children, &child{
		name:     name,
		start:    start,
		interval: interval,
	})
}

// Run starts every child under the given parent context, pulses the
// supervisor's own heartbeat, and returns it. The returned stop function
// cancels the supervisor's context, shutting the whole subtree down
// cooperatively.
func (s *Supervisor) Run(parentCtx context.Context, pulseInterval time.Duration) (heartbeat, func()) {
	ctx, cancel := context.WithCancel(parentCtx)
	s.mu.Lock()
	s.ctx = ctx
	s.cancel = cancel
	kids := make([]*child, len(s.children))
	copy(kids, s.children)
	s.mu.Unlock()

	// Give every child its own cancellable context derived from the supervisor.
	for _, c := range kids {
		cctx, ccancel := context.WithCancel(ctx)
		c.cancel = ccancel
		c.hb = c.start(cctx, c.interval)
		c.lastPulse.Store(time.Now().UnixNano())
		log.Printf("%s: started child %s", s.name, c.name)
	}

	hb := make(chan struct{}, 1)

	go func() {
		defer close(hb)

		ticker := time.NewTicker(pulseInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				select {
				case hb <- struct{}{}:
				default:
				}
			case <-ctx.Done():
				log.Printf("%s: stopped", s.name)
				return
			}
		}
	}()

	go s.supervise(kids)

	return hb, func() { cancel() }
}

// supervise watches every child's heartbeat and restarts unhealthy ones.
func (s *Supervisor) supervise(kids []*child) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
		}

		now := time.Now()
		var failed []*child

		s.mu.Lock()
		for _, c := range kids {
			var gotPulse, closed bool
			select {
			case _, ok := <-c.hb:
				gotPulse = true
				closed = !ok
			default:
			}
			if closed {
				// The child exited cleanly and closed its heartbeat.
				failed = append(failed, c)
			} else if gotPulse {
				c.lastPulse.Store(now.UnixNano())
			} else if now.Sub(time.Unix(0, c.lastPulse.Load())) > s.timeout {
				// Silence for too long: the child is wedged or dead.
				failed = append(failed, c)
			}
		}
		s.mu.Unlock()

		if len(failed) == 0 {
			continue
		}

		names := make([]string, len(failed))
		for i, f := range failed {
			names[i] = f.name
		}
		log.Printf("%s: unhealthy children %v (strategy %s)", s.name, names, s.strategy)

		switch s.strategy {
		case OneForOne:
			for _, c := range failed {
				s.restart(c)
			}
		case OneForAll:
			for _, c := range kids {
				s.restart(c)
			}
		case RestForAll:
			started := false
			for _, c := range kids {
				for _, f := range failed {
					if c == f {
						started = true
					}
				}
				if started {
					s.restart(c)
				}
			}
		}
	}
}

// restart cancels the old child's context and starts a fresh one in its place.
// The new heartbeat channel is installed under the lock before the old context
// is cancelled, so a concurrent supervision loop never observes a closed
// (and falsely "unhealthy") heartbeat.
func (s *Supervisor) restart(c *child) {
	s.mu.Lock()
	oldCancel := c.cancel
	ctx, cancel := context.WithCancel(s.ctx)
	c.cancel = cancel
	c.hb = c.start(ctx, c.interval)
	c.restarts.Add(1)
	c.lastPulse.Store(time.Now().UnixNano())
	s.mu.Unlock()

	oldCancel() // tell the old goroutine to stop now that its replacement is running

	log.Printf("%s: restarted %s (restart #%d)", s.name, c.name, c.restarts.Load())
}

// steadyWorker pulses forever on schedule, honouring context cancellation.
func steadyWorker(name string, _ time.Duration) startGoroutineFn {
	return func(ctx context.Context, pulseInterval time.Duration) heartbeat {
		hb := make(chan struct{}, 1) // buffered: a pulse is retained until read
		go func() {
			defer close(hb)
			logGoroutines(name, "running")
			ticker := time.NewTicker(pulseInterval)
			defer ticker.Stop()
			for {
				select {
				case hb <- struct{}{}:
				case <-ctx.Done():
					return // cooperatively exits
				}
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
		}()
		return hb
	}
}

// wedgedWorkerOnce pulses pulsesBeforeWedge times, then wedges forever
// (blocks in select {}). It wedges only once per process, so a replacement
// started by the supervisor is healthy.
func wedgedWorkerOnce(name string, pulsesBeforeWedge int) startGoroutineFn {
	var wedged atomic.Bool
	return func(ctx context.Context, pulseInterval time.Duration) heartbeat {
		hb := make(chan struct{}, 1) // buffered: a pulse is retained until read
		go func() {
			defer close(hb)
			logGoroutines(name, "running")
			ticker := time.NewTicker(pulseInterval)
			defer ticker.Stop()
			pulses := 0
			for {
				select {
				case hb <- struct{}{}:
				case <-ctx.Done():
					return // cooperatively exits (but never reached once wedged)
				}
				pulses++
				if pulses >= pulsesBeforeWedge && wedged.CompareAndSwap(false, true) {
					logGoroutines(name, "wedged after %d pulses; no longer pulsing", pulses)
					select {} // stuck forever: the fault a supervisor heals
				}
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
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

	// Halt everything after 6 seconds by cancelling the root context.
	ctx, cancelRoot := context.WithCancel(context.Background())
	time.AfterFunc(6*time.Second, func() {
		log.Println("main: halting supervision tree")
		cancelRoot()
	})

	// The supervision tree:
	//
	//     root (one_for_one, timeout 1 s)
	//     ├─ alpha   (steady worker, pulses 100 ms)
	//     ├─ beta    (sub-supervisor, one_for_one, timeout 400 ms)
	//     │   ├─ gamma  (steady worker, pulses 100 ms)
	//     │   └─ delta  (wedges after 3 pulses, pulses 100 ms)
	//     └─ epsilon (steady worker, pulses 100 ms)
	//
	// When delta wedges, beta notices the silence and restarts only delta.
	// root keeps pulsing the whole time, so main never sees the outage.

	root := NewSupervisor("root", 1*time.Second, OneForOne)

	beta := NewSupervisor("beta", 400*time.Millisecond, OneForOne)
	beta.Add("gamma", steadyWorker("gamma", 100*time.Millisecond), 100*time.Millisecond)
	beta.Add("delta", wedgedWorkerOnce("delta", 3), 100*time.Millisecond)

	root.Add("alpha", steadyWorker("alpha", 100*time.Millisecond), 100*time.Millisecond)
	root.Add("beta", func(ctx context.Context, interval time.Duration) heartbeat {
		hb, _ := beta.Run(ctx, interval)
		return hb
	}, 200*time.Millisecond)
	root.Add("epsilon", steadyWorker("epsilon", 100*time.Millisecond), 100*time.Millisecond)

	rootHB, stopRoot := root.Run(ctx, 200*time.Millisecond)

	for range rootHB {
		log.Println("main: heartbeat from root supervisor")
	}

	// rootHB closes when root's context is cancelled (after 6 s).
	stopRoot() // ensure the supervisor's context is cancelled
	logGoroutines("main", "done")
}
