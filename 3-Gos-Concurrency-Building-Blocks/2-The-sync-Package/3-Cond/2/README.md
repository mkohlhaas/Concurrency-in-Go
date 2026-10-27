# Better Names for `sync.Cond`

`sync.Cond` is one of Go's more cryptic names. It comes from "condition variable" (a concept from Hoare monitors, pthreads, and OS textbooks), but the name often trips people up.

## Why is it confusing?

- **It's not the condition itself.** `sync.Cond` is the rendezvous object that goroutines park on and get woken up from. The actual condition is whatever shared state you're guarding (queue length, a flag, "button clicked", etc.).
- **The API is about waiting, not testing.** `Wait()`, `Signal()`, and `Broadcast()` are about going to sleep until someone tells you to check again. You must always check the predicate in a `for` loop.
- **The mental model doesn't match.** In event-style patterns (like the button example in `cond-broadcast.go`), subscribers aren't waiting for `Clicked == true`—they're waiting for the *click to happen*. Calling it "a condition" feels backwards.

## Better alternative names

| Name | Notes |
|---|---|
| `ConditionVariable` | Most faithful to CS terminology. More verbose, but immediately recognizable if you've seen pthreads/monitors before. |
| `Event` | Great for "something happened" patterns. Methods like `Wait()`, `Fire()`/`Set()`, and `Broadcast()` read very naturally (used in Win32 and many other APIs). |
| `Notifier` | Emphasizes "notify waiters" over the condition variable formalism. Fits pub-sub style code well. |
| `Gate` | Nice for "block until opened" semantics, but less general purpose for arbitrary predicates. |

## Mental model

Treat `*sync.Cond` as a **waiting room**, not a condition. Goroutines go to the waiting room with `Wait()`, and `Signal()`/`Broadcast()` tells some or all of them "it's worth checking outside again".

## What does this code demonstrate?

This example shows using `sync.Cond` with `Broadcast()` to implement a simple **event/pub-sub pattern** where a single event wakes up *all* registered listeners at the same time.

### Step by step

1. **The event is represented by a condition variable**: `Button.Clicked` is a `*sync.Cond` backed by a mutex. Instead of a boolean flag, the "click" itself is the event that will notify waiters.
2. **Subscribers are registered**: `subscribe()` spawns a new goroutine for each listener. That goroutine locks the condition's mutex, calls `c.Wait()` (which blocks until notified), and then executes its callback once woken up.
3. **Synchronization ensures no one misses the broadcast**: `goroutineRunning` is used to block `subscribe()` until the new goroutine is about to call `Wait()`. This guarantees all three listeners are parked and waiting before anything is broadcast.
4. **The event is fired**: `button.Clicked.Broadcast()` wakes *every* goroutine currently blocked on `Wait()` simultaneously. Each one reacquires the lock, exits `Wait()`, runs its callback (printing a different message), and releases the lock.
5. **Completion is tracked**: A separate `clickRegistered` WaitGroup waits for all three callbacks to finish before `main` exits.

### Key lessons

- **`Broadcast()` vs `Signal()`**: `Signal()` would only wake a single waiter, while `Broadcast()` wakes all of them. This is exactly what's needed here, as all three UI actions need to react to the same click.
- **Condition variables are for notification, not state**: There is no predicate being checked in a `for` loop here. The "condition" is simply "the button was clicked", and the callbacks are executed as a reaction to that notification. This is why this pattern maps better to "events" than to "conditions".

## Go reality

Brevity wins in the stdlib (`Mutex`, `WaitGroup`, `Once`, `RWMutex`). `Cond` has been this way since Go 1.0 and is deeply embedded in books and codebases. If I were designing it today, I'd pick `ConditionVariable`, but it's more about learning the mental model than the name being fixable.
