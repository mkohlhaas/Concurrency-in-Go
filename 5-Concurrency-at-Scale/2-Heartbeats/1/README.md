# Heartbeat with work results

A small Go program demonstrating the classic **heartbeat + results** concurrency pattern:
a background worker emits regular "pulse" heartbeats to a consumer, and periodically
produces a work result, all while honoring a cancellation signal.

## What it does

- A `doWork` goroutine runs two tickers:
  - `pulse` — fires every `pulseInterval` (250 ms) and sends a heartbeat on the `heartbeat` channel.
  - `workGen` — fires every `2 * pulseInterval` (500 ms) and produces a work result (`time.Time`) on the `results` channel.
- The main loop `select`s over three channels:
  1. `heartbeat` — prints `pulse` on each beat.
  2. `results` — prints the seconds field of each result.
  3. `time.After(livenessTimeout)` — a 500 ms **liveness watchdog**. It is
     recreated on every loop iteration, so it only wins if no heartbeat or
     result arrives for a full 500 ms (i.e. the worker is stuck or dead). It is
     **not** an "exit after 500 ms" deadline: heartbeats arrive every 250 ms
     (`livenessTimeout/2`), resetting the watchdog each time. The program
     actually exits after `runForSeconds` (5 s) when `done` is closed.
- After `runForSeconds` (5) seconds, a `done` channel is closed, which shuts the worker down cleanly
  (both output channels are closed via `defer`).

## Output

```
pulse
pulse
results 18
pulse
pulse
results 18
...
```

Because Go's `select` picks randomly among ready cases and heartbeats are sent
non-blocking, the exact pulse/result interleaving — and even the counts — vary
from run to run.

## Files

- `interval-heartbeat.go` — the program source
- `go.mod` — module definition (`go 1.27.3`)

## Run it

```bash
go run .
```

It runs for about 5 seconds, then exits when `done` is closed.

## What happens after 5 seconds?

### Shutdown path diagram

```text
done closes
    ↓
worker returns (via <-done case)
    ↓
heartbeat & results channels closed (via defer)
    ↓
main loop select sees closed channel → !ok → return
```

### How it works

1. `time.AfterFunc` closes `done` after `runForSeconds`.
2. The worker goroutine sees `<-done` become ready, returns, and its `defer close(heartbeat)` / `defer close(results)` run — both output channels are closed.
3. Receiving from a closed channel returns immediately with `ok = false`, so both the `heartbeat` and `results` cases in the main loop's `select` are instantly ready.
4. Go's `select` picks one at random, sees `!ok`, and hits `return` — the program exits.

The `livenessTimeout` watchdog is **not** involved in this normal shutdown path.
It's only a fallback for when the worker dies without closing the channels.

## Notes

- `time.Tick` was replaced with `time.NewTicker` + `defer Stop()` so the tickers
  can be released cleanly (the old `time.Tick` can never be stopped).
- `go vet` and `go run -race` both pass — there are no data races.
- Remaining timing behaviour (by design of this example): pulses can be silently
  dropped if the main loop isn't receiving at that instant, and a result in flight
  can be lost when `done` closes.
