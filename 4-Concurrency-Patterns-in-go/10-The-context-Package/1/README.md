# greeter-with-done-chan

Demonstrates goroutine cancellation using the manual `done chan struct{}` pattern
— the pre-`context` idiom for signalling goroutines to stop.

## What it does

`main()` launches two concurrent tasks — a greeting and a farewell — each
tracked by a `sync.WaitGroup`. Both tasks call `locale(done)`, which simulates
a 5-second lookup blocked on a `select`:

* `<-done` — cancellation was signalled → returns a `"canceled"` error
* `time.After(5s)` — work completes → returns `"EN/US"`

A background goroutine closes `done` after 2 seconds, cancelling both lookups
before they finish. The `"canceled"` error propagates up through
`genGreeting`/`genFarewell` → `printGreeting`/`printFarewell` → `main`, which
prints:

```
greeting: canceled
farewell: canceled
both tasks finished
```

## The point

This shows how cancellation was done manually before Go's `context` package:
passing a `done` channel through every function, selecting on `<-done`, and
manually propagating errors. The next step in this series replaces this pattern
with `context.WithCancel`, `ctx.Done()`, and `context.Canceled`.

## Run

```sh
go run greeter-with-done-chan.go
```
