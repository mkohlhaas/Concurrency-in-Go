# Heartbeat + results — test fix (directory 4)

This directory contains a `DoWork` producer and a test for it, plus a benchmark.
It is part of the Chapter 5 heartbeat sequence (see `../1`, `../2`, `../3`).

## What it does

`DoWork(done, nums...)` runs a goroutine that:

1. Simulates **startup work** with `time.Sleep(2 * time.Second)` before the first value.
2. Sends a best-effort "pulse" heartbeat on the `heartbeat` channel before each number
   (buffered, cap 1, non-blocking — if the consumer hasn't read the previous pulse,
   this one is dropped rather than blocking the worker).
3. Sends each `int` on `intStream`, participating in a `select` with `<-done` so
   cancellation interrupts delivery.
4. Closes both output channels via `defer` when the loop finishes or `done` closes.

## The problem

The original `TestDoWork_GeneratesAllNumbers` threw the heartbeat channel away:

```go
_, results := DoWork(done, intSlice...)
```

and then read each value with a fixed 1 second deadline:

```go
case <-time.After(1 * time.Second):
    t.Fatal("test timed out")
```

Because `DoWork` sleeps for 2 seconds before producing the first value, the first
deadline always expired. The test could not tell **"producer is still starting up"**
from **"producer is dead"** — it had no liveness signal.

There was also no benchmark, so `./runtest.sh` (`go test -bench=.`) had nothing to run,
and `package main` with no non-test source failed to compile under `go test`
("could not import main").

## What changed

- `package main` → `package main_test` so `go test` compiles an external test package.
- `any` → `struct{}` on the signal channels (`done`, `heartbeat`), the idiomatic
  Go choice for channels that carry no payload.
- The test now captures the heartbeat and waits for the first pulse before draining
  results — confirming the producer is alive despite the initial startup delay.
- Added `BenchmarkDoWork`, which recreates `DoWork` each iteration, drains the
  stream, and closes `done` to avoid goroutine leaks.

## The message

> Don't just read from a data channel and hope. Use a heartbeat channel to detect
> liveness, and use `select` with timeouts/cancellation so you don't block forever
> on a silent producer.

The heartbeat is the producer's way of saying **"I'm still here"** between values.
Without it, a consumer can't distinguish a slow producer from a dead one. With it,
you can `select` over `heartbeat`, `results`, and a liveness watchdog (`time.After`)
and decide what to do when silence lasts too long.

## Files

- `bad_concurrent_test.go` — the `DoWork` producer plus test and benchmark
- `go.mod` — module definition (`go 1.27.3`)
- `runtest.sh` — `go test -bench=.`
- `README.md` — this file

## Run it

```bash
./runtest.sh
```

or verbosely:

```bash
go test -bench=. -v
```

## How this version differs from `../3`

`../3` is the happy-path baseline: a worker that runs 10 iterations and closes its
channels when done, with no time-based machinery. Directory 4 adds:

- a **2 s startup delay** before the first value, exposing why a plain data-channel
  read with a fixed timeout breaks;
- a **benchmark** to exercise the producer under repetition;
- a test that uses the **heartbeat to establish liveness** before consuming results,
  matching the pattern `../1` demonstrates with its liveness watchdog.

## Notes

- `go vet` passes.
- Because heartbeats are non-blocking and `select` chooses randomly among ready
  cases, pulse/result interleaving and exact counts can vary from run to run.
- Closing `done` while `DoWork` is blocked sending a value cancels that delivery
  (the `<-done` case in the result `select` wins).