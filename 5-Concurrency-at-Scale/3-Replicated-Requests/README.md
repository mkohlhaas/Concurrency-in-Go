# Replicated Requests

## Question: Is `close(done)` needed? (line 45)

**Yes — `close(done)` is required.**

Removing it causes a `fatal error: all goroutines are asleep - deadlock!`

### Why

This is the classic replicated-requests / "or-request" pattern:

1. Launch 10 goroutines that each do simulated work.
2. Main reads the **first** result from the shared `results` channel (`firstResult := <-results`).
3. Main then `close(done)` to cancel the remaining 9 goroutines.
4. Main waits for all goroutines to finish with `wg.Wait()`.

The `done` channel is the cancellation signal. It is receive-only inside `doWork` (`done <-chan any`), so nothing sends on it — no risk of a "send on closed channel" panic. It is closed exactly once, from `main`, after the first result and before `wg.Wait()`.

### What happens without `close(done)`

After main takes the first answer, it stops reading from `results`. The remaining 9 goroutines continue working and reach the second `select`:

```go
select {
case <-done:
case result <- id:
}
```

- `results` is **unbuffered** and nobody is receiving anymore, so `result <- id` blocks forever.
- Those goroutines never call `wg.Done()` (via defer), so `wg.Wait()` blocks forever → deadlock.

### What `close(done)` does

Closing a channel makes every receive on it immediately ready (yielding the zero value). So when the remaining goroutines hit either `select`, the `<-done` case fires, they skip sending on `results`, return, call `wg.Done()`, and `wg.Wait()` completes cleanly.

### Verification

- With `close(done)`: all 10 goroutines finish, program prints `Received first answer from #7.` and exits normally.
- Without `close(done)`: Go panics with `all goroutines are asleep - deadlock!` at `wg.Wait()`.
