# RWMutex vs. Mutex benchmark

A small micro-benchmark (adapted from *Concurrency in Go*, ch. 3) that measures
how `sync.RWMutex`'s read-locking behaves compared to a plain exclusive lock as
the number of concurrent readers grows.

## Usage

```sh
go run ./...
```

The program prints a table with one row per reader count (1, 2, 4, ... 524,288)
and the elapsed time in each of two configurations.

## How it works

- **`producer`** acquires the *write* lock 5 times and calls `time.Sleep(0)`
  (a scheduler yield) while holding it. This simulates an occasional writer.
- **`observer`** acquires the lock once and releases it. This simulates a
  reader.
- **`test(count, mutex, rwMutex)`** starts one producer and `count` observers,
  waits for all of them, and returns the elapsed time.
- **`main`** loops over `count = 2^i` and runs two variants against the same
  `sync.RWMutex m`:

  | Column                   | Producer lock  | Observer lock        | Meaning                                   |
  | ------------------------ | -------------- | -------------------- | ----------------------------------------- |
  | `RWMutex (read lock)`    | write lock     | `m.RLocker()` (read) | The intended read-heavy use case          |
  | `Mutex (exclusive)`      | write lock     | write lock           | Baseline: `m` used as a plain mutex       |

## What it demonstrates

`RWMutex` is not automatically faster than `sync.Mutex`. In this benchmark:

- **Few readers (roughly 1–64):** the plain exclusive lock is clearly faster.
  The read-lock path has extra atomic and semaphore bookkeeping, and with so
  little contention that overhead is pure cost.
- **Many readers (tens of thousands and up):** the two are essentially tied,
  with `RWMutex` occasionally a few percent ahead — within run-to-run noise.
  Because the critical section is trivial, readers barely contend and the
  occasional writer keeps interrupting the read path, so read locks never get
  the large parallel-speedup one might expect.

A sample run:

```text
Readers  RWMutex (read lock)  Mutex (exclusive)
1        17.681µs             1.741µs
2        9.102µs              3.37µs
4        19.892µs             3.378µs
8        9.527µs              5.827µs
16       21.858µs             6.792µs
32       18.151µs             11.671µs
64       52.07µs              28.415µs
...
65536    22.114269ms          22.560004ms
131072   45.325614ms          45.073986ms
262144   91.547168ms          91.256591ms
524288   181.923152ms         185.92878ms
```

Numbers vary between runs (each row is a single untimed repetition), so treat
small differences as noise. The takeaway: only use `RWMutex` for genuinely
read-heavy workloads with infrequent writers, and measure — its overhead can
easily make it slower than a plain `sync.Mutex`.

## Notes

- Both variants reuse the same `m`. This is safe only because `test` returns
  after `wg.Wait()`, so the two runs never overlap.
- The integer range syntax (`for range 5`, `for i := range 20`) requires
  Go 1.22 or newer.
