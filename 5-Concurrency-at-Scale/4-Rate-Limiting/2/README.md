# Rate Limiting Example

## What it demonstrates

This code demonstrates **rate limiting** with Go's `golang.org/x/time/rate` package — specifically, a **shared token-bucket rate limiter** that throttles concurrent access to an API client.

### Key points

1. **One shared limiter across callers and methods**

   Both `ReadFile` and `ResolveAddress` draw from the same `rate.Limiter`, so the *combined* rate of both operations is capped at ~3 events/second — not 3 per method. The 20 concurrent goroutines are serialized and queued by the limiter.

2. **Blocking backpressure**

   Each method calls `a.rateLimiter.Wait(ctx)` before doing work. `Wait` blocks until a token is available, so callers wait their turn rather than immediately rejecting excess requests. Under load, log lines are spaced roughly ~333 ms apart after the initial burst.

3. **Context-aware limiting**

   `Wait` accepts a `context.Context`, so a caller can cancel or time out its wait if it is no longer willing to wait for a token.

### Code structure

- `APIConnection` holds a single `*rate.Limiter` initialized to **3 events per second** with a **burst capacity of 1**.
- `main` spawns 20 goroutines (10 `ReadFile`, 10 `ResolveAddress`) that share one `APIConnection`.
- `wg.Wait()` blocks until all goroutines finish.

### Takeaway

A single `rate.Limiter` is a simple, centralized way to protect an external service (or any shared resource) from being overwhelmed by concurrent goroutines. The trade-off is that it limits **total** throughput across everything using that limiter; separate limiters are needed for independent per-method or per-client limits.

## How to run

```bash
go run simple-rate-limit.go
```

Expected output shows timestamps spaced roughly 333 ms apart after the first burst:

```
00:00:00 ReadFile
00:00:00 ResolveAddress
00:00:00 ResolveAddress
00:00:33 ReadFile
00:00:33 ResolveAddress
00:01:06 ReadFile
...
Done.
```
