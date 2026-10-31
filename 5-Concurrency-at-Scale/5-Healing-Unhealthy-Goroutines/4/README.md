# Stacked Stewards — healing a wedged supervisor

Example 1 heals a ward. This example stacks two stewards so that the **supervisor itself** can be healed: the
outer steward watches the inner steward, and the inner steward watches the ward.

```text
main
 └── outer steward   timeout 1.5 s   (its "ward" is the inner steward)
      └── inner steward   timeout 0.5 s   <- wedges after 3 heartbeats
           └── ward   reports every 250 ms
```

The stack works because `newSteward` returns a `startGoroutineFn`: the outer steward neither knows nor cares that
the thing it supervises is itself a steward.

## The fault

`newStewardWithWedge` is `newSteward` plus a demo-only hook. After the inner steward sends each heartbeat to its
parent, the hook is asked whether to wedge it. On the third heartbeat it says yes, and the steward runs `select {}`:
its monitor loop blocks forever. It stops pulsing and stops watching its ward, but it never returns — so it never
closes its heartbeat channel either.

That is the failure an outer steward exists for. A wedged *ward* is caught by its own steward; a wedged *steward*
is invisible to the ward and can only be caught by something above it. Note that a panic is not healed here: there
is no `recover`, so a panic in either layer takes the process down. Only hangs and clean exits are recoverable
in-process.

Only the first inner steward is wedged (an `atomic.Bool` makes the hook fire once), so the replacement is healthy
and the demo ends in a recovered state.

## What happens

| Time | Event |
| --- | --- |
| 0.00 s | outer starts inner #1, inner #1 starts ward #1 |
| 0.75 s, 1.50 s, 2.25 s | inner #1 pulses to the outer; after the third it wedges |
| 2.25 s | ward #1 blocks trying to send its next heartbeat: the steward is no longer reading |
| 3.75 s | the outer has seen 1.5 s of silence: `outer #1: ward unhealthy; restarting` |
| 3.75 s – 5.25 s | the outer closes inner #1's `done`, then waits out its grace period; the wedged inner never reacts |
| 5.25 s | the outer starts inner #2, which starts ward #2 |

## What to look for

1. **`main` never notices.** The outer steward keeps pulsing every 500 ms through the whole outage: `main:
   heartbeat from outer steward` appears before, during and after the repair. Inner restarts are masked from above.
2. **The wedge is found by the clock, not the channel.** A crashed steward closes its heartbeat and is replaced at
   once; a wedged one stays silent and costs the full `outerTimeout`.
3. **Recovery costs two timeouts.** One to notice, one to stop waiting for a steward that will never stop. That
   bound (`awaitWardStop`) is what keeps the outer steward from being held hostage by its own ward.
4. **The leak is real.** Go cannot kill a goroutine. The wedged inner #1 and its blocked ward #1 are still running
   when `main` logs `goroutines = 3` at the end: main plus the two orphans. Stacking heals the *service*, not the
   leak — reclaiming those goroutines needs process-level supervision.

## Run it

```sh
go run .
```

It runs for 10 seconds, then `main` closes `done`.

## Sample output (abridged)

```text
07:57:49 main: starting (goroutines = 1)
07:57:49 outer #1: started; watching ward
07:57:49 inner #1: started; watching ward
07:57:49 ward #1: running (goroutines = 4)
07:57:49 ward #1: heartbeat 1
...
07:57:52 inner #1: wedged after 3 heartbeats; no longer pulsing (goroutines = 4)
07:57:52 main: heartbeat from outer steward
07:57:53 main: heartbeat from outer steward
07:57:53 outer #1: ward unhealthy; restarting (goroutines = 4)
07:57:55 outer #1: ward replaced
07:57:55 inner #2: started; watching ward
07:57:55 ward #2: running (goroutines = 6)
07:57:55 ward #2: heartbeat 1
...
07:57:59 main: halting outer steward
07:57:59 ward #2: halting
07:57:59 inner #2: ward stopped; returning
07:57:59 outer #1: ward stopped; returning
07:57:59 main: done (goroutines = 3)
```

Reading the counts:

| Count | Meaning |
| --- | --- |
| 4 | the healthy stack: main + outer + inner + ward |
| 6 | after the repair: the healthy stack **plus** the wedged inner #1 and its blocked ward #1 |
| 3 | at the end: main plus the two goroutines the wedge leaked |

`goroutines = 6` is the whole point of the example: the outer steward restored the service, and the cost was two
goroutines that can never be reclaimed from inside the process.
