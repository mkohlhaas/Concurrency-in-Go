# Heartbeat + results with an unhealthy worker

A small Go program demonstrating the classic **heartbeat + results** concurrency pattern:
a background worker emits regular "pulse" heartbeats to a consumer, and periodically
produces a work result, all while honoring a cancellation signal.

## What it does

- A `doWork` goroutine runs two tickers:
  - `pulse` — fires every `pulseInterval` (250 ms) and sends a heartbeat on the `heartbeat` channel (non-blocking; pulses can be dropped).
  - `workGen` — fires every `2 * pulseInterval` (500 ms) and produces a work result (`time.Time`) on the `results` channel.
- The worker loop runs **at most 10 iterations** (`for range 10`) and then returns. This simulates a worker that does some work and then becomes unhealthy (stops, crashes, or exits without closing its output channels).
- The main loop `select`s over three channels:
  1. `heartbeat` — prints `pulse` on each beat.
  2. `results` — prints the seconds field of each result.
  3. `time.After(livenessTimeout)` — a 500 ms **liveness watchdog**, recreated on every loop iteration. It only wins if no heartbeat or result arrives for a full 500 ms — i.e. the worker is stuck or dead.

## Shutdown sequence

The worker simulates becoming unhealthy by finishing its 10 iterations and returning.
It never closes its output channels, so main has no clean signal that the worker is gone.

### Step by step

1. **Worker runs.** The `for range 10` loop sends pulses every 250 ms and results every 500 ms until it completes 10 iterations, then returns.
2. **Worker cleanup.** On return, `defer pulse.Stop()` and `defer workGen.Stop()` release the tickers. **`heartbeat` and `results` are never closed** — there is no `close()` call on either channel anywhere in the program.
3. **Main blocks.** Once the worker stops sending, the `<-heartbeat` and `<-results` cases in main's `select` block forever (open channels, no senders). The `ok == false` path is never reached.
4. **Watchdog fires.** After `livenessTimeout` (500 ms) with no heartbeat or result, `time.After(livenessTimeout)` wins the `select`.
5. **Exit.** Main prints `worker goroutine is not healthy!` and returns. The program exits.

The `done` channel (closed at T+5s via `time.AfterFunc`) does not participate in this shutdown path — the watchdog trips well before 5 seconds elapse.

## Output

```
pulse
pulse
results 18
pulse
pulse
results 18
...
worker goroutine is not healthy!
```

Because Go's `select` picks randomly among ready cases and heartbeats are sent non-blocking, the exact pulse/result interleaving — and even the counts — vary from run to run.

## Notes

- `time.Tick` was replaced with `time.NewTicker` + `defer Stop()` so the tickers can be released cleanly (the old `time.Tick` can never be stopped).
- `go vet` and `go run -race` both pass — there are no data races.
- The liveness watchdog is the **actual** shutdown mechanism here. It reports the worker as "not healthy," which is accurate from main's perspective (silence for 500 ms), but the worker itself exited cleanly after 10 iterations — it just never signaled shutdown by closing its output channels.
- A result in flight can be lost if `done` closes while `sendResult` is blocked delivering it.

## Files

- `interval-heartbeat-misbehaving-goroutine.go` — the program source
- `go.mod` — module definition (`go 1.27.3`)

## Run it

```bash
go run .
```

It runs until the liveness watchdog trips (well under 5 seconds), then exits.
