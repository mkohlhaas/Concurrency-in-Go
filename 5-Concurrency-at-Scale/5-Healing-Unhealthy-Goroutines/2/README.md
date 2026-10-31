# Comparison: Ward vs More Complicated Ward

Adding a `results` channel is the main API change, but the real complexity is in *how the steward has to multiplex and redefine what "alive" means*. 

## The API Change

| Aspect | `../1/ward.go` | `../2/more-complicated-ward.go` |
|---|---|---|
| `startGoroutineFn` return | `(heartbeat <-chan struct{})` | `(heartbeat <-chan struct{}, results <-chan error)` |
| Steward return | `heartbeat <-chan struct{}` | `(heartbeat <-chan struct{}, results <-chan error)` |
| Responsibility | Only observe liveness | Observe liveness *and* forward work results upstream | 

## Why it's more complicated

| Concern | #1 (Simple Ward) | #2 (More Complicated Ward) |
|---|---|---|
| Liveness | Only heartbeats reset the timeout. If the heartbeat channel closes, it's treated as an explicit exit and the ward is restarted immediately. If nothing comes in for `timeout`, it's restarted. | *Either* a heartbeat or a received result resets the timeout (both `continue monitorLoop`, which recreates `time.After(timeout)`). A ward can now be "healthy" by producing errors, not just pulses. |
| Multiplexing | Selects over pulse, ward heartbeat, timeout, and done. | Selects over pulse, ward heartbeat, ward results, timeout, and done. On `wardResults` it must *forward* `err` via `results <- err` (blocking send), so backpressure propagates to the steward's caller while still needing to keep monitoring liveness. |
| Failure detection | Checks `ok` when reading heartbeats (to detect a closed channel). Has `awaitWardStop()` with a grace period to avoid interrupting a ward mid-shutdown and to prevent it from holding the steward hostage. | No `ok` checks for `wardHeartbeat` or `wardResults`. If either channel is closed, a receive of the zero value will fire in the select and loop forever (a busy loop) unless something else happens. It also only restarts on timeout, not on a clean ward exit. |
| Shutdown | Explicitly kills the ward, drains/awaits its stop with grace, and logs state transitions. | Closes `wardDone` if it exists and defers closing both channels, but is much less defensive about draining in-flight messages or handling a ward that has already exited. | 

## The takeaway

The extra channel isn't just "return another value" - it changes the health model to activity-based liveness, forces the steward to be a proxy (not just an observer), and introduces much easier ways to busy-loop or mishandle closed channels. #1 is more verbose because it's explicitly handling those edge cases; #2 is shorter but cuts a lot of that robustness. That's the real jump in complexity.

## Why is the channel named "results" although it's an error channel?

It's named by its *purpose* in the pipeline, not by the Go type it carries.

### Why "results" makes sense here

- **Role-based naming**: This follows the classic worker pattern of `jobs` and `results`. The channel represents the stream of work *outcomes* from the ward, not just "errors". In a real implementation, those results could be `struct{Value T, Err error}`, success values, or any other output.
- **Extensibility**: Naming it `results` keeps the API generic. If this example were expanded, you could send non-error values without having to rename the channel or change the steward's forwarding logic. The type (`error`) is just a simplification for this toy example.
- **Separation from liveness**: Heartbeats answer "is it alive?" while "results" answer "what did it produce?". These are two different concerns, so they get two different channels with names that reflect their roles, even if one of them only produces failure cases here.

### The quirk

In this specific code, the ward only ever sends `fmt.Errorf(...)` and never a `nil` to indicate success, so it's effectively an error channel in practice. A clearer name like `errors` or `failures` would be less misleading for this exact example, but "results" is used to demonstrate the pattern of adding an output stream rather than strictly describing the element type.

## Can you build a stable production system this way?

Yes — but you have to stop writing code the way this example does. This tutorial code is deliberately *not* production-grade. It's built to demonstrate a pattern, not to be robust.

The fickery comes from habits that work in learning examples but are deadly in production:

| What we did here | Why it's fickle | Production alternative |
|---|---|---|
| `time.Sleep(1s)` outside a `select` | Can't be interrupted; shutdown signals queue up and may be missed | `select { case <-time.After(...): case <-ctx.Done(): }` |
| Raw `done <-chan struct{}` channels | No deadlines, no propagation, easy to forget to close | `context.Context` with `ctx.Done()` |
| `time.Sleep(2s)` in main to wait for ward | Just hopes the goroutine gets scheduled in time | `sync.WaitGroup`, `errgroup.Group`, or channel-based coordination |
| Blocking channel sends inside `select` | If nobody reads, the steward or ward freezes forever | Buffered channels, non-blocking sends with `default`, or separate result goroutine |
| Assuming `ok == false` means clean exit | A closed channel just means no more values; the sender may have panicked | Explicit error handling, `recover` in ward, log stack traces |

### What a stable version looks like

```go
func doWork(ctx context.Context) (<-chan struct{}, <-chan error) {
    hb := make(chan struct{}, 1)  // buffered so sends don't block
    errs := make(chan error, 1)

    go func() {
        defer close(hb)
        defer close(errs)

        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return  // clean, immediate shutdown
            case <-ticker.C:
                select {
                case hb <- struct{}{}:
                default:  // nobody listening, don't block
                }
            }
        }
    }()

    return hb, errs
}
```

### The real answer

Stable production systems are built from **small, well-tested primitives**, not clever steward architectures. Use `context.Context` for cancellation, `golang.org/x/sync/errgroup` to manage goroutine groups, never use `time.Sleep` to synchronize shutdown, and write tests that inject failures deterministically rather than hoping timing works out. Keep the surface area small: one goroutine does one thing, owns its channels, and shuts down when its context is canceled.
