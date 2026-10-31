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
