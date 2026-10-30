# Heartbeats

Implements a `DoWork` function that streams numbers on `intStream` while
periodically signaling a `heartbeat` channel on its own interval, honoring a
cancellation `done` channel.

### Key behavioral difference

In `../5`, every emitted number is preceded by a heartbeat signal, so the
heartbeat rate is tied to the number production rate.

In `../6`, the heartbeat is driven by an independent ticker (`time.Tick`).
Numbers flow through `intStream` at the consumer's pace (the send only
succeeds when the receiver is ready), while heartbeat ticks fire on `pulseInterval`
regardless of how fast numbers are consumed.

## Running

```
go test -v -count=1 .
go test -bench=.        # same as runtest.sh
```

The test takes ~3s: `DoWork` sleeps 2s before its first heartbeat pulse, then all numbers stream back-to-back once the first heartbeat is received — within the test's 2s-per-iteration timeout.
