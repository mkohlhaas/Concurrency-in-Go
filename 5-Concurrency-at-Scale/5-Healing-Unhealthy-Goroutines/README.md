# Healing Unhealthy Goroutines — the Steward Pattern

This directory contains four progressively more advanced examples of the **Steward Pattern**: a concurrency design pattern that monitors a worker goroutine (the "ward") and automatically restarts it if it becomes unresponsive or crashes.

## The examples

| Directory | What it demonstrates |
|---|---|
| `1/` | The canonical steward pattern — complete, correct, and fully documented. |
| `2/` | Steward with a results channel — multiplexes liveness heartbeats and work results. |
| `3/` | A compact, incomplete steward — minimal timeout-based healing only; not production-grade. |
| `4/` | Stacked stewards — an outer steward heals a wedged inner steward. |

## Should example `3/` go in `1/`?

**No.** Directory `3/` is shorter in lines of code, but it is *less complete and less robust*, so it is not a better "most basic" example.

| | `1/ward.go` | `3/example-steward-usage.go` |
|---|---|---|
| Lines | ~250 + full README | ~121, no README |
| Types | Named types (`heartbeat`, `doneChan`, `startGoroutineFn`) | Inline `func(...)` with `any`, no named types |
| Detects a crash | Yes — closed heartbeat channel (`ok == false`) restarts immediately | No — ward returns `nil` heartbeat, so `<-wardHeartbeat` blocks forever; only a full timeout detects failure |
| Waits for ward to stop | Yes — bounded `awaitWardStop()` grace period | No — closes `wardDone` and immediately restarts |
| Ward sends heartbeats | Yes — proves liveness | No — explicitly does not send heartbeats |
| Timer | `time.NewTicker` + `defer Stop()` | `time.Tick` (leaks — cannot be stopped) |
| Bundled unrelated code | No | Includes an `or` fan-in/done helper (a different pattern) |
| Documentation | Comprehensive README with sequence diagram & "why the details matter" | None |

The short `3/` example omits the safety mechanisms that make the steward pattern correct: closed-channel crash detection, bounded ward-stop waiting, a heartbeat-sending ward, and stoppable tickers. It also bundles in an unrelated `or` fan-in function. Directory `1/` is the right place to start.

## Intended reading order

`1 → 2 → 3 → 4`:

1. **`1/`** — learn the pattern correctly (the steward watches a heartbeat-sending ward, detects crashes and deadlocks, waits for wards to stop on shutdown).
2. **`2/`** — add an output channel so the steward also forwards work results upstream; liveness becomes "heartbeat *or* result" activity.
3. **`3/`** — a minimal, deliberately incomplete variant that only heals via timeout. Useful for seeing how small the code *can* be, but not a model to copy.
4. **`4/`** — stack stewards so an outer steward heals a wedged inner steward (heal the supervisor itself).

Each directory is a self-contained `go run .` example with its own `go.mod`.

## Running an example

```sh
cd 1
go run .
```

## Production note

These are teaching examples, not production code. Use `context.Context` for cancellation, `golang.org/x/sync/errgroup` for goroutine groups, never `time.Sleep` to synchronize shutdown, and test failure injection deterministically rather than relying on timing.