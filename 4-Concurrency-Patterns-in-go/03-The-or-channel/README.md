# The OR Channel

This program demonstrates the **OR channel** pattern in Go — a way to wait for
the first of several concurrent operations to complete.

## What it does

The `or` function takes a variadic list of channels and returns a single new
channel. That returned channel becomes ready as soon as **any one** of the input
channels becomes ready — the concurrency equivalent of the boolean `||` operator.

In other words: instead of waiting for all of several channels to close, you wait
for the first one that does.

## How it works

1. **Zero channels** → returns `nil` (reading from a `nil` channel blocks forever, which is standard behavior for an empty OR — there's nothing to complete).
2. **One channel** → passes it through unchanged.
3. **Two channels** → a single `select` waits on either one; whichever fires
   lets the goroutine finish, which closes the returned `orDone` channel.
4. **Three or more channels** → the goroutine selects over the first three
   channels, and for the rest it recursively calls `or`, appending its own
   completion channel (`orDone`) to the tail so the recursion terminates.

The caller receives `orDone`, which closes when the first source signals.

The example wires up five timer-based signals (`sig`) with different delays and
feeds them into `or`. Since the 1-second signal is the earliest, the program
completes after ~1 second and prints how long it took.

## Try it

```bash
go run or-channel.go
```

Expected output (timing will vary):

```
Done after ~1.0s.
```
