# Errgroup — error-propagating goroutine groups

This example demonstrates **`golang.org/x/sync/errgroup`**, an Errgroup-style
goroutine group that collects errors and (optionally) cancels a context on the
first failure.

It is written for someone who has never seen `errgroup` before. Each example is
self-contained and heavily commented.

## What errgroup is

`errgroup.Group` is like `sync.WaitGroup`, but with two additions:

1. **It returns the first non-nil error** from any goroutine started with
   `g.Go()`.
2. **`errgroup.WithContext`** derives a context that is **cancelled when the
   first goroutine returns an error**. Every other goroutine can observe
   `ctx.Done()` and stop early.

Without a context, `errgroup` behaves like `sync.WaitGroup` + an error bucket.
With a context, errors propagate as cancellation — not just as a return value.

## The four examples

| Example | What it shows |
|---|---|
| **1** | Basic `errgroup.Group`: launch N goroutines, wait for all, collect the first error |
| **2** | `errgroup.WithContext`: one worker fails → context cancelled → other workers stop via `ctx.Done()` |
| **3** | Fan-in / scatter-gather: one worker fails → watcher goroutine stops early via `ctx.Done()` |
| **4** | Supervisor under errgroup: a restart loop gives up after N failures, returns a fatal error, cancels the context, and shuts down a sidecar goroutine |

## Errgroup vs supervisor

Example 5 in this series implements a **supervisor tree**: a failed child is
*restarted* so the service keeps running.

`errgroup` takes the opposite philosophy: if any goroutine fails, the whole
group is *cancelled*. It is best for "all-or-nothing" workloads — fan-out,
scatter-gather, map-reduce, or any operation where one failure means the whole
thing aborts.

The two patterns can be combined: run a supervisor inside an errgroup, so that
when healing is exhausted the supervisor returns a fatal error, cancels the
errgroup context, and shuts down everything else (example 4 above).

## Run it

```sh
cd 6
go run .
```

You will need network access once, so `go mod tidy` can download
`golang.org/x/sync`.

## Relationship to the other examples

| Directory | Pattern | Failure model |
|---|---|---|
| `4/` | Stacked stewards | A wedged steward is restarted by the steward above it |
| `5/` | Supervision tree | Multiple children, restart strategies (`one_for_one`, etc.) |
| `6/` | Errgroup | First error cancels the whole group via context |

Each directory is a self-contained `go run .` example with its own `go.mod`.
