# Error Handling in Concurrency — `checkStatus` Pattern

## What this example does

`checkStatus` fans out concurrent `http.Get` requests over a channel and
collects the successful responses. Failed requests (e.g. bad hosts) are
logged and skipped rather than crashing the program.

## Changes made (compared to the original code)

1. **Replaced the manual `done` channel with `context.Context`.**
   The original used a hand-rolled `<-chan any` closed via `defer close(done)`
   as a cancellation signal.

2. **Promoted `checkStatus` from an anonymous function inside `main` to a
   named function** that accepts a `context.Context` parameter.

3. **Added explicit cancellation checks.**
   The worker now selects on `<-ctx.Done()` both before issuing a request and
   before handing a response off on the channel, instead of selecting on the
   hand-rolled done channel.

4. **Removed the unsafe cancellation pattern.**
   The original used `defer close(done)` in `main` to close the done channel.
   Closing a channel twice panics in Go, and deferring a close inside `main`
   is the wrong place for it anyway (it runs only when `main` returns, long
   after the work is done). Replaced with the idiomatic
   `context.WithCancel(context.Background())` + deferred `cancel()`.

5. **Expanded the URL list.**
   Originally: `["https://www.google.com", "https://badhost"]` (plus a
   commented-out alternative). Now:
   - `https://www.google.com`
   - `https://www.github.com`
   - `https://httpbin.org/delay/3` (slow endpoint)
   - `https://www.example.com`

6. **Added mid-point cancellation.**
   `main` now counts received responses and calls `cancel()` after the second
   response, causing the worker to stop fetching the remaining URLs.

## Notes

- The `responses` channel is unbuffered; each send blocks until the consumer
  receives it, and the worker closes it when finished so `range` terminates.
- `http.Get` does not accept a context; the worker checks `ctx.Done()` before
  each request. For real-world cancellation of in-flight requests, use
  `http.NewRequestWithContext` with a context-aware client.
