# Deadlock with sync.Once (Circular Dependency)

This example demonstrates a **deadlock** caused by creating a circular dependency between two `sync.Once` instances.

## Code Walkthrough

```go
var onceA, onceB sync.Once

// Forward declaration
var initB func()

initA := func() { onceB.Do(initB) }
initB = func() { onceA.Do(initA) }

onceA.Do(initA)
```

## Why It Deadlocks

1. `onceA.Do(initA)` starts execution. `onceA` marks itself as "in progress" and acquires its internal lock while running `initA`.
2. `initA` calls `onceB.Do(initB)`. `onceB` marks itself as "in progress" and runs `initB`.
3. `initB` calls `onceA.Do(initA)`. `onceA` is already in-progress and locked by the original call, so `onceA.Do()` blocks waiting for `initA` to complete.
4. Meanwhile, `initA` is blocked waiting for `initB` to finish (it called `onceB.Do(initB)`). Neither goroutine can proceed → **circular wait / deadlock**.

## Key Takeaway

`sync.Once` is not safe from deadlocks when initialization functions form a cycle. The `Do` method holds an internal lock while executing the function, so re-entering `Do` on the same `Once` instance from within its own initialization function will block forever.
