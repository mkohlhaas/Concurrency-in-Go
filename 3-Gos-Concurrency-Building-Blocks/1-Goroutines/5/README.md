# Measuring the Size of a Goroutine

> Goroutines are cheap, but they are not free. This program measures how much
> memory one costs.

`goroutine-size.go` spawns a large number of goroutines, keeps them alive, and
compares process memory before and after. The result is a rough per-goroutine
figure — typically a couple of kilobytes.

## How it works

1. **`memConsumed`** forces a `runtime.GC()` and then reads `runtime.MemStats.Sys`,
   the total memory obtained from the OS. The GC makes the reading stable so the
   "before" and "after" numbers are comparable.

2. **The goroutines** all run the same `noop` function, which calls `wg.Done()` and
   then blocks forever on `<-c`. `c` is a **nil channel**, and a receive from a nil
   channel never completes, so every goroutine stays parked.

3. **The measurement** happens after `wg.Wait()` returns but while the goroutines
   are still blocked, so the difference between `memBefore` and `memAfter`
   attributes the new memory to them.

## The goroutines never finish — on purpose

This is the subtle part, and the reason the code is written the way it is:

- `wg.Done()` is called **before** `<-c`. The WaitGroup therefore only tracks
  "the goroutine has started and reached its parking spot", not "the goroutine
  is finished". That is what lets `wg.Wait()` return while the goroutines are
  still alive for the measurement.
- The goroutines themselves never return. They cannot be woken: `c` is a
  receive-only nil channel, nothing is ever sent, and it cannot be reassigned.
  They are terminated only when the process exits at the end of `main`.
- This is a deliberate goroutine leak, and a blocked goroutine is a live GC
  root, so it can never be collected either. `runtime.GC()` in `memConsumed`
  does not clean them up; it only stabilizes the heap before reading `Sys`.
- In a long-running program this pattern would leak roughly 2 KB per goroutine
  permanently. The usual fix is a channel you close (or a `context`) so the
  goroutines can exit:

  ```go
  done := make(chan struct{})
  noop := func() { <-done }
  // ... spawn, measure ...
  close(done) // let every goroutine return
  ```

  Note that a `sync.WaitGroup.Go`-based version cannot express this demo: `Go`
  calls `Done` only after the function returns, and `noop` never returns, so
  `Wait` would deadlock. That is why the explicit `wg.Add`/`wg.Done` form is used.

## Running

```console
$ go run goroutine-size.go
```

## Expected output

```text
2.530KB
```

The exact number varies with Go version, architecture, and how much the process
happens to allocate around the measurement. It is an order-of-magnitude figure,
not a precise contract.

## Caveats

- `MemStats.Sys` counts all memory obtained from the OS, not just goroutine
  stacks, so the result understates nothing but overstates a little.
- The measurement assumes `memAfter > memBefore`; the subtraction is on `uint64`,
  so a shrinking process would wrap around.
- Units are bytes divided by `1024`, printed as `KB`.

## Further reading

- [Concurrency in Go](https://katherine.cox-buday.com/concurrency-in-go/) by Katherine Cox-Buday
- [`runtime.MemStats` — Go standard library](https://pkg.go.dev/runtime#MemStats)
- [`sync.WaitGroup` — Go standard library](https://pkg.go.dev/sync#WaitGroup)
- [Go 1.25 release notes — `WaitGroup.Go`](https://go.dev/doc/go1.25)
