# sync.Cond in Go

> A condition variable is a low-level synchronization primitive that lets goroutines
> sleep and wait for a specific state change.

## Overview

`sync.Cond` implements a condition variable. Unlike channels, which transfer data
payloads, a `sync.Cond` sends no data — it acts purely as a signalling system, a
"waiting room" for goroutines. It is paired with an underlying lock (a `sync.Mutex`
or `sync.RWMutex`) to safely manage access to the shared state.

## How it works

Every `sync.Cond` relies on three methods to coordinate execution:

| Method        | Behaviour                                                                                    |
| ------------- | -------------------------------------------------------------------------------------------- |
| `Wait()`      | Suspends the current goroutine and adds it to the condition's internal notification queue.   |
| `Signal()`    | Wakes up exactly one goroutine suspended in the queue. A no-op if no goroutines are waiting. |
| `Broadcast()` | Wakes up **all** goroutines currently waiting on the condition.                              |

### What `Wait()` does to the lock

1. Calling `Wait()` automatically **unlocks** the associated mutex, so other
   goroutines can change the state.
2. When the goroutine is eventually woken up, `Wait()` automatically **re-locks**
   the mutex before returning.

## The canonical usage pattern

To use `sync.Cond` correctly and avoid race conditions or deadlocks, always check
the state condition in a `for` loop rather than an `if` statement.

```go
mu.Lock()
defer mu.Unlock()

for !ready {
    cond.Wait() // suspends, unlocks mu, re-locks mu when woken
}

// Safe: the condition still holds and mu is held.
```

A complete, runnable example lives in [`main.go`](./main.go):

```console
$ go run .
Worker: Condition met, processing data...
```

## Why must you check `Wait()` inside a `for` loop?

Writing `if !ready { cond.Wait() }` introduces a critical vulnerability due to two
phenomena:

- **Spurious wakeups** — a goroutine can occasionally wake up without an explicit
  `Signal()` or `Broadcast()` being sent by your code.
- **State hijacking** — if `Broadcast()` wakes five goroutines, they all contend
  for the re-locked mutex one by one. By the time the third or fourth goroutine
  acquires the lock, a faster goroutine may already have consumed the resource or
  flipped `ready` back to `false`.

The `for` loop ensures that, upon waking, a goroutine re-evaluates the state and
only proceeds when it is truly safe to do so.

## `sync.Cond` vs. channels

Channels are heavily favoured for general concurrency in Go because they are safer
and more idiomatic. `sync.Cond` excels in niche, performance-critical patterns:

| Feature                 | Channels (`chan`)                                                                 | `sync.Cond`                                                                        |
| ----------------------- | --------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| **Data flow**           | Transfers ownership or data payloads.                                             | Signalling only; no data is passed.                                                |
| **Repeated broadcasts** | Closing a channel acts as a one-time broadcast — it cannot be reopened or reused. | `Broadcast()` can be called repeatedly over the lifetime of the application.       |
| **Memory overhead**     | Allocating thousands of individual channels can consume substantial memory.       | Uses a highly optimised internal `notifyList` (linked list) with minimal overhead. |

### Rule of thumb

Use `sync.Cond` when you need a **reusable broadcast event** where multiple workers
wake up dynamically based on a complex shared condition, or when writing
ultra-high-performance system constructs such as ring buffers or memory pools.

## Further reading

- [`sync.Cond` — Go standard library](https://pkg.go.dev/sync#Cond)
- [Share Memory By Communicating — The Go Blog](https://go.dev/blog/codelab-share)
