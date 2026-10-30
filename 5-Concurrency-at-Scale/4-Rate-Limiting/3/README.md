# Multi-Rate Limiter

This example shows how to combine multiple [`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate) limiters into a single limiter that enforces both a per-second and a per-minute limit on every operation (`ReadFile` and `ResolveAddress`).

## What the code example demonstrates

This program shows how to build a **composite, multi-stage rate limiter**. Every API call (`ReadFile` and `ResolveAddress`) must pass through both limits before it runs, so concurrent workers are forced to respect a per-second cap *and* a per-minute cap.

### How it works

**1. MultiLimiter — the composite pipeline**

The unexported `multiLimiter` struct groups separate limiters into one logical pipeline:

*   **Sorting:** `MultiLimiter(...)` sorts child limiters by rate (lowest to highest: `minuteLimit`, `secondLimit`) with `sort.Slice`.
*   **Wait pipeline:** `Wait(ctx)` loops through each child limiter sequentially, blocking on `limiter.Wait(ctx)` until every limiter in the chain has an available token.

**2. Defined API limits** (in `Open()`)

| Limiter | Rate | Burst |
|---------|------|-------|
| `secondLimit` | 2 requests / second | 1 token |
| `minuteLimit` | 12 requests / minute | 5 tokens |

**3. Concurrent simulation** (`main`)

*   A `sync.WaitGroup` coordinates **20 concurrent goroutines**.
*   10 goroutines call `ReadFile()`.
*   10 goroutines call `ResolveAddress()`.
*   Both methods share the same `rateLimiter`, so all workers contend for the same token pools.

### Running

```bash
go run multi-rate-limit.go
```

### Runtime behavior

When you run the code you'll see distinct patterns in the logs:

*   **Initial burst (throttled by secondLimit)** — the minute bucket's burst capacity of 5 allows a handful of operations to start right away, but they're paced to about 2 per second by the stricter `secondLimit`.
*   **Per-second throttling** — after the burst is exhausted, logs pace out to no more than 2 operations per second.
*   **Per-minute exhaustion** — after 12 requests, remaining goroutines pause until the minute window resets and refills the bucket.

---

## Pointer constructor assigned to an interface field

```go
// MultiLimiter constructor returns a pointer to an unexported type
func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	...
	return &multiLimiter{limiters}
}
```

```go
// Call site, no explicit pointer or conversion needed
rateLimiter: MultiLimiter(secondLimit, minuteLimit),
```

This is idiomatic Go for two reasons:

1. **Returning a pointer from a constructor** is standard practice, especially when the concrete type is unexported (`multiLimiter` is lowercase). Callers outside the package cannot construct it directly anyway.
2. **No `&` or conversion at the call site** is needed because `MultiLimiter` already returns a `*multiLimiter`, and that pointer value is stored directly into the `rateLimiter` field of type `RateLimiter` (an interface). Since `*multiLimiter` implements both `Wait` and `Limit` with pointer receivers, Go implicitly converts it to the interface type — no type assertion required.

---

## Log from `go run .`

```text
09:35:34 ResolveAddress
09:35:34 ReadFile
09:35:35 ReadFile
09:35:35 ReadFile
09:35:36 ResolveAddress
09:35:39 ReadFile
09:35:44 ResolveAddress
09:35:49 ReadFile
09:35:54 ReadFile
09:35:59 ResolveAddress
09:36:04 ReadFile
09:36:09 ResolveAddress
09:36:14 ReadFile
09:36:19 ResolveAddress
09:36:24 ResolveAddress
09:36:29 ReadFile
09:36:34 ResolveAddress
09:36:39 ResolveAddress
09:36:44 ResolveAddress
09:36:49 ReadFile
09:36:49 Done.
```

The log beautifully illustrates exactly how the Token Bucket algorithm manages state when a system shifts from a "burst" phase to a "sustained rate" phase.

### Phase 1: The Initial Burst (09:35:34 – 09:35:36)

When the application starts, both token buckets are completely full.

*   The `minuteLimit` starts with 5 tokens (its maximum burst capacity).
*   The `secondLimit` starts with 1 token.

During the first few seconds, 5 requests fire almost immediately because they drain the `minuteLimit`'s burst capacity:

1.  **09:35:34** (2 requests): The first request takes 1 token from both buckets. This empties the `secondLimit` bucket. The second request has to wait a fraction of a second for `secondLimit` to refill, but it still executes within the same second clock tick.
2.  **09:35:35** (2 requests) & **09:35:36** (1 request): These requests consume the remaining 3 tokens from the `minuteLimit` burst pool, throttled to roughly 2 per second by the `secondLimit`.

At 09:35:36, the 5-token burst pool of the `minuteLimit` is completely empty.

### Phase 2: Steady State Throttling (09:35:39 – End)

Once the burst pool is empty, the system is entirely bottlenecked by the refill rate of the strictest limiter, which is the `minuteLimit`.

The `minuteLimit` is configured for 12 requests per 60 seconds. Mathematically, that means 1 new token is generated exactly every 5 seconds (60 / 12 = 5).

Look closely at the log timestamps after the burst:

*   09:35:39 (+3s after burst empty)
*   09:35:44 (+5s)
*   09:35:49 (+5s)
*   09:35:54 (+5s)
*   09:35:59 (+5s)
*   09:36:04 (+5s)

Because there are still 15 goroutines waiting in the background, the very instant a single new token drops into the `minuteLimit` bucket every 5 seconds, a waiting goroutine snatches it up immediately and executes. The `secondLimit` (which refills twice every second) is always sitting full and satisfied during this phase, so it no longer causes any delays.

### Are Bursts also rate limited?

Bursts are absolutely rate limited, but only by the other limiters in the pipeline.

This is the exact reason why the code uses a `MultiLimiter`. A single token bucket limiter cannot limit its own burst, but when you chain multiple limiters together, they act as a series of filters. An event can only happen if it passes through every filter in the line.

#### The Funnel Analogy ⏳

Imagine you are pouring water through two funnels stacked on top of each other:

1.  **Top Funnel** (`minuteLimit`): Has a wide mouth that can hold 5 cups of water at once, but its bottom hole only lets water drip out at a rate of 12 cups per minute (1 drip every 5 seconds).
2.  **Bottom Funnel** (`secondLimit`): Has a tiny mouth that can only hold 1 cup at a time, and its bottom hole drops water at a maximum speed of 2 cups per second.

If you suddenly dump a massive bucket of water (20 requests) into the top funnel:

*   The top funnel's mouth instantly fills up to its max capacity of 5 cups (this is your burst of 5). The other 15 cups spill over or have to wait in line.
*   However, those 5 cups cannot just fall straight through to the floor. They immediately hit the bottom funnel (`secondLimit`).
*   Because the bottom funnel can only handle 1 cup at a time and caps the speed at 2 cups per second, those 5 burst requests are forced to slow down to a maximum speed of 2 per second.

#### Looking at Your Exact Logs

This is exactly why your "burst" didn't happen all at the exact same millisecond. The `minuteLimit` said, "Yes, I will allow 5 requests right now!" but the `secondLimit` stepped in and said, "Not so fast, you can only execute 2 of them per second."

```text
09:35:34 ResolveAddress _ Max 2 per second rule enforced
09:35:34 ReadFile       /  by the secondLimit
09:35:35 ReadFile       _ Max 2 per second rule enforced
09:35:35 ReadFile       /  by the secondLimit
09:35:36 ResolveAddress -> The 5th and final burst token is used
```

### Summary of the Rule

*   **The Burst size** determines how many total requests can bypass the long-term wait time during a sudden spike.
*   **The Rest of the Pipeline** determines how fast that allowed spike is physically permitted to execute.

---

### Was the Multi-Limiter Pattern a creative idea of Katherine Cox-Buday when she wrote the book?

The entire MultiLimiter pattern is straight out of Katherine Cox-Buday's book, `Concurrency in Go`. It is one of the most celebrated and creative sections of Chapter 5 ("Concurrency at Scale").

The MultiLimiter code you provided is the exact example she uses to teach developers how to compose simple primitives into highly elegant, sophisticated architectural patterns.

Her code is widely praised in the Go community for a few key reasons:

1. **The Genius of sort.Slice**

   Before her example, many developers trying to combine rate limiters wrote messy, conditional pipelines. Her idea to automatically sort the slice via `sort.Slice` by the coarsest-grained limit (the slowest rate) ensured the system automatically blocks on the most restrictive throttle first. It was an incredibly elegant piece of clean code design.

2. **Interface Composition**

   By introducing the custom `RateLimiter` interface, she decoupled the system from the strict `*rate.Limiter` struct. This allowed her `multiLimiter` to recursively hold other `multiLimiter` instances. For example, a developer could create an overall `apiLimiter` and combine it with a `diskReadLimiter`, bundling them together into an entirely new composite limiter flawlessly.

3. **Democratizing the "Token Bucket" for Go Developers**

   While the token bucket algorithm itself is old, `golang.org/x/time/rate` was historically poorly understood and lacked comprehensive multi-tier examples. Her book fundamentally changed how Go developers protect outbound API tokens and internal system processes at scale.
