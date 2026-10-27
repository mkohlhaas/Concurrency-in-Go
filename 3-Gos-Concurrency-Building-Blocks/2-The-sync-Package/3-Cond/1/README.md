# A Bounded Queue with `sync.Cond`

> [`cond-based-queue.go`](./cond-based-queue.go) builds a bounded queue out of a
> condition variable instead of a channel. `main` is the producer, and every
> enqueue spawns a consumer that removes an item one second later.

```console
$ go run .
Adding to queue: #1
Adding to queue: #2
Adding to queue: #3
Removed from queue: #1
...
```

## The cast

| Piece | Role |
| --- | --- |
| `queue []struct{}` | the shared buffer (elements carry no data, only the length matters) |
| `cond := sync.NewCond(&sync.Mutex{})` | condition variable plus the mutex that guards `queue` |
| the `for range capacity` loop | producer: 20 iterations, one enqueue each |
| `removeFromQueue` | consumer: sleep → lock → pop → print → `Signal` |
| `consumers sync.WaitGroup` | tracks in-flight consumers so `main` can wait for them |
| `added, removed int` | running counts for the log lines, both incremented under `cond.L` |

`cond.L` is just the mutex inside the cond. Both sides lock *the same* mutex,
which is what makes `Wait` and `Signal` work.

## The `sync.Cond` API

The whole API is five things: `NewCond`, the `L` field, and three methods.

### `sync.NewCond(l Locker) *Cond`

Constructor. `l` is the mutex that guards the condition's state:

```go
cond := sync.NewCond(&sync.Mutex{}) // or &sync.RWMutex{}
```

`l` must be non-nil, and every goroutine that reads or changes the guarded
state must use the same lock.

### `cond.L`

The `Locker` you passed in, stored as a field. You lock and unlock through it:

```go
cond.L.Lock()
defer cond.L.Unlock()
```

It exists so `Wait` can unlock and re-lock *your* mutex internally — that is how
the cond and your state stay in sync.

### `cond.Wait()`

Suspends the calling goroutine until it is signalled. Requires `cond.L` to be
held, otherwise it panics.

Under the covers it does this **atomically**:

1. adds the goroutine to the cond's internal FIFO notify list,
2. unlocks `cond.L`,
3. parks the goroutine,
4. on wake-up, re-locks `cond.L`,
5. returns.

Because of step 4, when `Wait` returns you hold the lock again and can safely
re-read the state. Because of step 3, the wake-up only means "something
changed" — never "your condition is now true".

### `cond.Signal()`

Wakes **exactly one** waiter — the one that has been waiting longest (FIFO). If
nobody is waiting, it is a no-op; the signal is not remembered. It does not
block and does not release any lock.

### `cond.Broadcast()`

Wakes **all** current waiters. They then contend for the mutex one at a time,
so they resume serially, not simultaneously. Also a no-op with no waiters, and
also not remembered.

### Rules that matter

| Rule | Why |
| --- | --- |
| Always hold `cond.L` when calling `Wait` | It panics otherwise, and it needs the lock to release it safely |
| Always wrap `Wait` in a `for` predicate check | Wake-ups can be spurious, and another goroutine may have consumed the condition before you got the lock |
| Change state while holding the lock, then signal | Keeps the predicate true at the moment the waiter re-checks it |
| Prefer `Signal` if one waiter can proceed; use `Broadcast` when any/all may be able to | `Broadcast` causes a thundering herd of lock contention |
| Never assume a signal arrives | Signal and Broadcast are edge-triggered with no memory — if you are not waiting, you missed it |
| Do not copy a `Cond` after use | It contains a `noCopy` marker; `go vet` flags it |
| `Wait` is not cancellable | No `context` support, unlike `select` on a channel. If you need cancellation, a channel is usually the better tool |

A minimal shape tying it together:

```go
cond.L.Lock()
defer cond.L.Unlock()

for !condition() { // re-check after every wake-up
    cond.Wait() // unlocks, parks, re-locks
}
// here the condition holds and the lock is held
```

In this program the producer does exactly this (`for len(queue) > maxQueued`),
and a consumer does `queue = queue[1:]` then `Signal()` — one waiter, so one
signal is enough.

## Step by step

**Producer**:

```go
for range capacity {          // 20 iterations (Go 1.22+ integer range)
    cond.L.Lock()             // enter critical section

    for len(queue) > maxQueued {
        cond.Wait()           // full? sleep, releasing the lock
    }

    queue = append(queue, struct{}{})
    added++
    fmt.Printf("Adding to queue: #%d\n", added)
    consumers.Add(1)                  // register before the goroutine starts
    go removeFromQueue(time.Second)   // schedule a consumer
    cond.L.Unlock()                   // leave critical section
}

consumers.Wait()   // outside the lock: consumers need cond.L to finish
```

`Wait()` atomically releases `cond.L` and parks the goroutine. When it is
woken, it re-acquires the lock before returning, so the queue is protected
again by the time the `for` condition is re-checked. That re-check loop is
mandatory — you cannot trust that the state is still valid on wake-up.

**Consumer**:

```go
defer consumers.Done()
time.Sleep(delay)     // 1s, outside the lock — doesn't block the producer
cond.L.Lock()
defer cond.L.Unlock()
queue = queue[1:]     // pop the front
removed++
fmt.Printf("Removed from queue: #%d\n", removed)
cond.Signal()         // wake the producer if it is waiting
```

`Signal` wakes exactly one waiter. Here only the producer ever waits, so
`Signal` is enough (`Broadcast` would also work).

## What actually happens

Instrumenting a copy to log elapsed time and queue length gives:

```
[ 0.00s] Adding to queue: #1  (len: 1)
[ 0.00s] Adding to queue: #2  (len: 2)
[ 0.00s] Adding to queue: #3  (len: 3)
[ 1.00s] Removed from queue: #1  (len: 2)
[ 1.00s] Removed from queue: #2  (len: 1)
[ 1.00s] Removed from queue: #3  (len: 0)
...
[ 6.01s] Removed from queue: #18 (len: 2)
[ 7.01s] Removed from queue: #19 (len: 1)
[ 7.01s] Removed from queue: #20 (len: 0)
[ 7.01s] main returns
```

| Time | What happens | `len(queue)` after |
| --- | --- | --- |
| 0.00s | first three enqueues; each spawns a consumer that sleeps 1s. The producer then blocks in `Wait`. | 3 |
| 1.00s | the three consumers wake as a batch, pop and `Signal`; the producer refills to 3. | 3 |
| 2.00s–5.00s | the burst repeats roughly once per second; items 7–18 are enqueued. | 3 |
| 6.00s | removals free slots; items 19–20 are enqueued and the producer loop ends. | 2 |
| 7.01s | the last consumers pop the final items and call `Done`; `consumers.Wait()` returns and `main` exits. | 0 |

Because `main` waits, all 20 `"Adding"` and all 20 `"Removed"` lines are printed.

## Points worth noticing

- **`maxQueued` is not the buffer size.** The producer waits only while
  `len(queue) > 2`, so the queue can transiently hold **3** items.
  `capacity = 20` is just the slice's initial capacity — and is re-used as the
  number of items to produce.
- **`main` waits for its consumers.** `consumers.Add(1)` happens before each
  goroutine starts, so `Wait` cannot miss it, and `Wait` is called *outside*
  the critical section — a consumer needs `cond.L` before it can finish, so
  waiting while holding the lock would deadlock.
- **Throughput is bursty** because every consumer sleeps the identical 1s.
  Stagger the delays and the output interleaves more smoothly.
- **The counters need no atomics.** `added` and `removed` are only read and
  written while holding `cond.L`, so plain `int`s are safe. If the counts were
  read by a goroutine outside the lock (say, a stats printer), they would need
  `sync/atomic` or their own mutex.
- **All access happens under `cond.L`** — the slice is never touched without
  it, which is the invariant the example rests on.
- **`Wait` must be called with the lock held**, otherwise it panics.

So it is a hand-rolled bounded queue whose blocking behaviour comes from a
condition variable rather than from channel buffer semantics — the textbook
illustration of when `sync.Cond` is useful.

## Is `struct{}{}` idiomatic?

Yes — though the use in this example is a niche, if legitimate, pattern.

`struct{}` is a type with zero fields, so it has a size of zero bytes.
`struct{}{}` is a value of that type. It means "a value that carries no
information" — it exists purely to occupy a slot or to signal an event.

### Where it is clearly idiomatic

```go
set := map[string]struct{}{} // set semantics: no value needed, no allocation
done := make(chan struct{})  // "close means done"; a value cannot be sent
```

Both are extremely common in real Go code, and `chan struct{}` is the standard
signal channel. Using `bool` for a set, or `chan bool` for a signal, is the less
idiomatic choice.

### In this example

The elements genuinely carry no data — only `len(queue)` matters — so
`struct{}` is honest and costs nothing. It is the standard trick for "counted
slots" without allocating. Two caveats worth knowing:

1. **Slice, not a set.** `[]struct{}` is not the idiomatic *set* form; the map
   form is. A slice of empty structs is a legitimate but niche pattern.
2. **`queue = queue[1:]` never reclaims the head.** It re-slices, so the front
   of the backing array is abandoned until `append` outgrows `cap` and
   reallocates. Harmless for a ten-element demo, but not how you would write a
   long-lived queue.

### Alternative: a counter is more honest

Since the elements are meaningless, a plain counter says the same thing with
less machinery:

```go
var (
    queued int
    cond   = sync.NewCond(&sync.Mutex{})
)

// producer
for queued > maxQueued {
    cond.Wait()
}
queued++

// consumer
queued--
cond.Signal()
```

This drops both the slice and the `queue[1:]` re-slicing, and makes the `Cond`
predicate the star of the demo. The trade-off: you lose the visual "this is a
queue of items" framing, which is presumably why the book used a slice in the
first place.

## Further reading

- [The empty struct — Dave Cheney](https://dave.cheney.net/2014/03/25/the-empty-struct)
- [`sync.Cond` — Go standard library](https://pkg.go.dev/sync#Cond)
- [The Go Memory Model](https://go.dev/ref/mem)
