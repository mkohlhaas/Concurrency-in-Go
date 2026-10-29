# Heartbeat + results (minimal version)

A small Go program demonstrating the classic **heartbeat + results** concurrency pattern:
a background worker emits regular "pulse" heartbeats to a consumer and periodically
produces a work result, all while honoring a cancellation signal.

## What it does

- `doWork` spawns a goroutine that runs 10 iterations. On each iteration it:
  1. Sends a best-effort "pulse" on the `heartbeatStream` channel (buffered, cap 1,
     non-blocking — if the consumer hasn't read the previous pulse, this one is dropped
     rather than blocking the worker).
  2. Sends a random integer (`rand.Intn(10)`) on the `workStream` channel. The send
     participates in a `select` with `<-done`, so cancellation interrupts delivery.
- When the loop finishes, both output channels are closed via `defer`.
- `main` `select`s over both channels, printing `pulse` on each heartbeat and `results <n>`
  on each result. It exits as soon as either channel is closed (`ok == false`).
- `done` is a `chan struct{}` — a pure signal channel (no payload). It is closed by
  `main` via `defer`, but in this example nothing ever triggers that close, so the
  program exits naturally after the worker's 10 iterations complete.

## Output

```
pulse
results 7
pulse
results 1
...
```

10 pulses interleaved with 10 results (order varies between runs because `select`
picks randomly among ready cases, and heartbeats can be silently dropped).

## Files

- `work-unit-pulse.go` — the program source
- `go.mod` — module definition (`go 1.27.3`)
- `README.md` — this file

## Run it

```bash
go run .
```

## How this version differs from `../1` and `../2`

| | `../1` (`interval-heartbeat.go`) | `../2` (`interval-heartbeat-misbehaving-goroutine.go`) | `../3` (this one) |
|---|---|---|---|
| **Trigger mechanism** | `time.Ticker` (time-based pulses and work) | `time.Ticker` (time-based pulses and work) | Plain `for range 10` loop — no timers at all |
| **Work payload** | `time.Time` | `time.Time` | `int` (`rand.Intn(10)`) |
| **Worker lifecycle** | Runs until `done` is closed (after 5 s via `time.AfterFunc`) | Runs 10 iterations then returns; **never closes output channels** | Runs 10 iterations then returns; **closes both channels via `defer`** |
| **Main exit paths** | `done` close, closed channel, or liveness watchdog | Liveness watchdog trips after 500 ms of silence | Either output channel closes |
| **Liveness watchdog** | Yes (`time.After(livenessTimeout)`) | Yes — detects the silent/misbehaving worker | **None** |
| **`done` channel** | Closed after 5 s | Closed after 5 s (never participates in shutdown) | Created but never closed in this run |

### The key distinction

`../3` is the **minimal baseline**. It removes the time-based machinery (`time.Ticker`,
`time.AfterFunc`, liveness watchdog) entirely to isolate the core channel pattern:
non-blocking heartbeat, cancellable result delivery, and clean shutdown through channel
closing.

`../1` adds time-based triggering and a proper timed shutdown. `../2` demonstrates the
failure mode where a worker stops without closing its output channels, forcing `main`
to rely on a liveness watchdog instead. `../3` shows the happy path where the worker
finishes and signals completion by closing its channels.

## Notes

- `go vet` passes — no data races.
- Because heartbeats are non-blocking and `select` chooses randomly among ready cases,
  pulse/result interleaving and exact counts vary from run to run.
- A result in flight can be lost if `done` closes while the worker is blocked sending it.
