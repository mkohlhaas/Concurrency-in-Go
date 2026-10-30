# no-rate-limit.go

Sample code from *Concurrency in Go* by Katherine Cox-Buday, chapter 5: Concurrency at Scale → Rate Limiting (example 1).

## What it's trying to demonstrate

**The problem of unbounded concurrency — calling an external API with no rate limiting.**

The program opens an `APIConnection` and launches 20 goroutines all at once:

* 10 goroutines call `ReadFile`
* 10 goroutines call `ResolveAddress`

A `sync.WaitGroup` waits for all of them to finish. The two API methods are just stubs (`// Pretend we do work here`), so there's no simulated delay and no throttling whatsoever.

## The point

All 20 requests fire simultaneously, with no coordination, no token bucket, no leaky bucket, and no cap on how many in-flight calls can hit the service at once. In a real program this could overwhelm the downstream API, exceed connection/concurrency limits, or trip the remote side's own rate limiter (causing errors or HTTP 429s).

Because this is example "1" in the rate-limiting section, its purpose is to show the naive, unthrottled approach so that the following examples can contrast it with rate-limited patterns (typically built on buffered channels and tickers).

**In short:** it demonstrates what happens when you don't rate limit — a burst of 20 concurrent calls with no control — setting up the need for rate limiting.
