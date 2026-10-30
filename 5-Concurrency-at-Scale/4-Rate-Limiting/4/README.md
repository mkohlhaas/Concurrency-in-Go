# Tiered Rate Limiter

This example shows how to **tier** rate limits so that different operations enforce different combinations of limits — for example, API calls plus disk I/O for `ReadFile`, and API calls plus network I/O for `ResolveAddress`.

It builds on the composite limiter idea from `../3` (multi-rate-limit.go) and adds **recursive composition**: each tier is itself a `MultiLimiter`, and those tiered limiters are then combined again per operation.

## What the code example demonstrates

The program runs 20 concurrent goroutines against an `APIConnection`:

*   10 goroutines call `ReadFile()`
*   10 goroutines call `ResolveAddress()`

Before an operation runs it must acquire a token from **every limiter in its tier pipeline** (`multiLimiter.Wait` loops through each child limiter sequentially and blocks until all have a token).

### How it works

**1. Tiered limiters** (`Open()`)

Each tier is a composite `MultiLimiter` of one or more `*rate.Limiter`s:

| Tier | Composition | Rate | Burst |
|------|-------------|------|-------|
| `apiLimit` | MultiLimiter | 2 / sec **and** 12 / min | 2 **and** 10 |
| `diskLimit` | MultiLimiter (single limiter) | 1 / sec | 1 |
| `networkLimit` | MultiLimiter (single limiter) | 3 / sec | 3 |

**2. Per-operation pipelines** (`ReadFile` / `ResolveAddress`)

| Operation | Must pass through | Reason |
|-----------|-------------------|--------|
| `ReadFile` | `apiLimit` → `diskLimit` | API calls plus disk I/O |
| `ResolveAddress` | `apiLimit` → `networkLimit` | API calls plus network I/O |

**3. `multiLimiter` — recursive composite**

The unexported `multiLimiter` struct implements the `RateLimiter` interface (`Wait` + `Limit`), so a `multiLimiter` can be nested inside another `MultiLimiter`. Children are sorted by rate (slowest/strictest first) so the pipeline blocks on the most restrictive throttle first.

### Running

```bash
go run tiered-rate-limit.go
```

## Comparison to `../3` (multi-rate-limit.go)

| | `3/multi-rate-limit.go` | `4/tiered-rate-limit.go` |
|---|---|---|
| **Structure** | One shared `rateLimiter` field | Three tier fields: `apiLimit`, `diskLimit`, `networkLimit` |
| **Composition depth** | Flat — `MultiLimiter` holds `*rate.Limiter`s directly | Tiered — each tier is itself a `MultiLimiter` |
| **Operation limits** | Both operations use the **same** combined limiter | Each operation uses a **different** tier combination |
| **Recursive** | No | Yes — `multiLimiter` can hold other `multiLimiter`s |

In short: version `3` shows how to **combine** multiple limits into one pipeline that applies to everything. Version `4` shows how to **tier** limits — build reusable composite limiters per resource (API/disk/network), then **mix and match** them per operation.

## Interface composition

The custom `RateLimiter` interface (`Wait(context.Context) error` + `Limit() rate.Limit`) decouples the system from `golang.org/x/time/rate`'s `*rate.Limiter`. Because `*multiLimiter` implements that interface with pointer receivers, it can be passed as a variadic argument to `MultiLimiter(...)` and stored into interface fields with no explicit conversion — enabling the recursive, tiered design.
