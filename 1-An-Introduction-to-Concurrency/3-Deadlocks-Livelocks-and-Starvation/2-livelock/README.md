# Live-lock in a Hallway

> A live-lock is when goroutines keep changing state in response to each other,
> but make no forward progress.

This Go program simulates a classic live-lock scenario between two people (Alice
and Megan) trying to pass each other in a narrow hallway. Unlike a deadlock,
where programs completely freeze, a live-lock happens when multiple active
goroutines continuously react to one another without ever getting anywhere.

## Key takeaway

Alice and Megan constantly step side to side in perfect unison, blocking each
other until they give up after three attempts. Because the program depends on
Go's scheduler, one of them may occasionally slip past instead.

"Success!" represents the rare case where the live-lock breaks and someone
actually gets through; it's the scheduler, not output ordering, that decides it.

## How it works

1. **The ticking clock (`cadence`)** — a background goroutine uses `sync.Cond`
   to broadcast a signal every millisecond. This acts like a metronome, forcing
   Alice and Megan to move at exactly the same pace.

2. **The shared variables (`left` and `right`)** — these track how many people
   are trying to move to that side of the hallway. They are modified with
   `sync/atomic` to avoid data races.

3. **The mirror dance (`tryDir`)** — a person steps to a side (say, left),
   incrementing the counter. Because they are perfectly synchronized, Alice and
   Megan both increment `left` at the same time. Each then checks whether the
   value is `1`: if it were, the path would be clear. But since both stepped
   left, the value is `2`. Seeing the path blocked, they both step back,
   decrement the counter, and immediately try the other direction (right).

4. **The exasperation** — being perfectly in sync, they repeat this polite
   dodge-dance three times, fail to pass each other, and print that they "toss
   their hands up in exasperation!"

## Running

```console
$ go run livelock-hallway.go
```

## Expected output

The exact interleaving depends on Go's scheduler, so the two walkers' lines can
appear in either order. A typical run looks like this:

```text
Alice is trying to scoot: left right left right left right
Alice tosses her hands up in exasperation!
Megan is trying to scoot: left right left right left right
Megan tosses her hands up in exasperation!
```

Now and then the scheduler lets one person through, in which case that line gains
a `. Success!` suffix and the walker does not toss her hands up.

## Are Alice and Megan walking concurrently?

Yes — but with an important qualifier: they run concurrently, yet not
independently.

**Concurrent: yes.** `main` starts two separate goroutines:

```go
go walk(&peopleInHallway, "Alice")
go walk(&peopleInHallway, "Megan")
```

Both are alive at the same time, both call `tryDir`/`takeStep`, and both contend
on the same shared `left`/`right` counters (via `atomic`) and the same `cadence`
condition variable. `peopleInHallway.Wait()` only returns when both have
finished. They are genuinely concurrent — Go concurrency, independent of how
many cores you have.

**But they are not independent — they move in lockstep.** `takeStep` is a
barrier paced by the metronome: every goroutine inside `Wait()` is released by
the same `Broadcast` every 1 ms. So Alice and Megan advance together, one
synchronized step per tick, rather than walking freely. A goroutine that arrives
late simply misses a tick and waits for the next one, which re-synchronizes it.

That lockstep is the live-lock. Each walker calls `atomic.AddInt32(dir, 1)`
*before* its `takeStep()` and reads the counter *after* it. Because both
increments land before the shared tick, both checks see `2` instead of `1`, so
neither gets `. Success!`; both back off (`atomic.AddInt32(dir, -1)`) and retry.
A run confirms it — six directional attempts each, no success:

```text
Megan is trying to scoot: left right left right left right
Megan tosses her hands up in exasperation!
...
```

One thing not to misread: seeing Megan's line before Alice's does not mean Megan
ran first. Each goroutine accumulates into its own `bytes.Buffer` and prints it
in a `defer` when it exits, so the print order only reflects which goroutine
finished first. The actual interleaving of their steps happens inside the 1 ms
ticks and is not visible in this output.

Strictly speaking, parallel execution on multiple cores is possible but not
guaranteed; the concurrency is guaranteed by the goroutines. The live-lock
itself only needs concurrency.

## Further reading

- [Concurrency in Go](https://katherine.cox-buday.com/concurrency-in-go/) by Katherine Cox-Buday
- [`sync.Cond` — Go standard library](https://pkg.go.dev/sync#Cond)
- [`sync/atomic` — Go standard library](https://pkg.go.dev/sync/atomic)
