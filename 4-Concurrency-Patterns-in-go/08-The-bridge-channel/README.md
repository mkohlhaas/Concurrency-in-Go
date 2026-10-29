# Bridge Channel

This file demonstrates the **bridge channel** concurrency pattern: flattening a
channel of channels (`<-chan <-chan T`) into a single stream of values
(`<-chan T`).

## The pattern

`bridge(done, chanStream)` consumes a stream of inner channels and merges their
values onto one output channel. It processes each inner channel **serially** —
the current channel is drained completely before the next one is picked up — so
values are never interleaved across concurrent inner channels.

Cancellation is cooperative: whenever the external `done` channel is closed,
`bridge` stops reading, stops forwarding, and closes its output channel. When
`chanStream` itself is closed, the output is closed too.

## How it works

1. Wait for either `done` or the next inner channel from `chanStream`.
2. Drain that inner channel completely, forwarding each value to the output.
3. Repeat until `chanStream` is closed or `done` is closed.
4. Close the output channel so `range` over it terminates.

The cancellation-aware forwarding helper (`orDone`) handles the per-value
`done` checks and closed-channel detection; those details are covered elsewhere.

## Output

```
0 1 2 3 4 5 6 7 8 9
```

## Run it

```sh
go run bridge-channel.go
```

## Notes

- Written with Go generics (`[T any]`), so it works for any element type.
- Only depends on the standard library (`fmt`).
- The `bridge` pattern serializes inner-channel consumption; if you need
  concurrent draining of many inner channels, a fan-out/fan-in approach is
  more appropriate.