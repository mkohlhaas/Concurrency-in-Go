# Channels

A channel is a typed conduit through which goroutines send and receive values.
It both moves data and synchronizes the goroutines on either end, so no explicit
locks are needed for the hand-off.

Go follows the Communicating Sequential Processes (CSP) model:

> Do not communicate by sharing memory; instead, share memory by communicating.

## Core characteristics

- **Typed** — a channel carries values of exactly one type, e.g. `chan int` or
  `chan string`.
- **FIFO** — values are received in the order they were sent.
- **Blocking by default** — sends and receives block until the other side is
  ready, which is what provides the synchronization.
- **Reference type** — like slices and maps, a channel must be initialized with
  `make` before use; the zero value is `nil`, and operations on it block forever.
- **Closeable** — `close` signals that no more values will be sent. It does not
  flush or clear the channel, and only the sending side should close it.

## Basic syntax

```go
// Create a channel with make.
ch := make(chan string)

// Send: the arrow points into the channel.
ch <- "Hello Go"

// Receive: the arrow points out of the channel.
message := <-ch

// Close: signal that no more values are coming.
close(ch)
```

## Types of channels

How a channel behaves depends on whether it has a buffer.

| Type | Declaration | Behavior | Typical use |
| --- | --- | --- | --- |
| Unbuffered | `make(chan int)` | Synchronous. The sender blocks until a receiver takes the value, and vice versa. | Strict synchronization and immediate coordination. |
| Buffered | `make(chan int, 3)` | Asynchronous up to the capacity. Senders block only when the buffer is full; receivers block only when it is empty. | Decoupling producer and consumer speeds, worker pools, rate limiting. |

### Directional channels

Function signatures can restrict a channel to send-only (`chan<- T`) or
receive-only (`<-chan T`). This is a compile-time safety measure and documents
who is allowed to write to or close the channel.

```go
func producer(out chan<- int) { /* may only send */ }
func consumer(in <-chan int)  { /* may only receive */ }
```

## Channel state behavior

The same operation can succeed, block, or panic depending on the channel's
state. This table is worth memorizing:

| Operation | `nil` channel | Open channel | Closed channel |
| --- | --- | --- | --- |
| Send `ch <- v` | Blocks forever | Succeeds, or blocks if full | Panics |
| Receive `<-ch` | Blocks forever | Succeeds, or blocks if empty | Returns the zero value; `ok == false` |
| Close `close(ch)` | Panics | Succeeds | Panics |

A `nil` channel is therefore not useless: it can be used in a `select` to
disable a case permanently.

## Receiving from a closed channel

The comma-ok idiom reports whether the received value came from an open channel:

```go
val, ok := <-ch
if !ok {
    fmt.Println("Channel is closed and empty!")
}
```

Receiving from a closed channel never blocks. It drains any buffered values
first, then yields the element type's zero value with `ok == false`. A `range`
loop over a channel does the same check for you and stops when the channel is
closed:

```go
for val := range ch {
    fmt.Println(val)
}
```

## Under the hood: `hchan`

Channels are not simple pipes. A channel variable is a pointer to a runtime
structure called `hchan` (defined in `runtime/chan.go`), which contains:

- **A circular buffer** — the queue that holds values for a buffered channel.
- **A mutex** — protects all fields in `hchan` from concurrent access. Channels
  do use locks internally, so callers do not have to.
- **Wait queues (`sendq` and `recvq`)** — linked lists of blocked goroutines
  waiting to send to a full or unbuffered channel, or to receive from an empty
  one.

### Direct copy

When a sender blocks because no receiver is ready, it is parked in `sendq`. If a
receiver then arrives, the runtime copies the value directly from the sender's
stack to the receiver's stack. This bypasses the circular buffer entirely and
avoids an unnecessary allocation.

## Multiplexing with `select`

To wait on several channels at once, use `select`. It behaves like a `switch`
whose cases are channel operations, blocking until one of them can proceed. If
multiple cases are ready at the same time, Go chooses one at random, which keeps
the selection fair:

```go
select {
case msg1 := <-ch1:
    fmt.Println("Received from ch1:", msg1)
case ch2 <- "Data":
    fmt.Println("Sent data to ch2")
default:
    // Runs immediately if every other case would block.
    fmt.Println("No channel was ready")
}
```

A `default` case makes the `select` non-blocking. See
[`4-The-select-Statement`](../4-The-select-Statement) for a fuller treatment.

## Examples in this directory

Each numbered directory is a standalone module; run one with `cd <n> && go run .`.

| Example | File | Demonstrates |
| --- | --- | --- |
| 1 | [`simple-chan.go`](./1/simple-chan.go) | An unbuffered channel: a goroutine sends, `main` receives. |
| 2 | [`chan-recv-multi-value.go`](./2/chan-recv-multi-value.go) | The two-value receive form `value, ok := <-ch`. |
| 3 | [`reading-from-closed-channel.go`](./3/reading-from-closed-channel.go) | Receiving from a closed channel: zero value, `ok == false`. |
| 4 | [`iterating-over-channel.go`](./4/iterating-over-channel.go) | `range` over a channel until the sender closes it. |
| 5 | [`unblocking-goroutines.go`](./5/unblocking-goroutines.go) | Closing a `chan struct{}` as a broadcast signal to waiting goroutines. |
| 6 | [`using-buffered-chans.go`](./6/using-buffered-chans.go) | A buffered channel with a producer and consumer. |
| 7 | [`chan-ownership.go`](./7/chan-ownership.go) | Channel ownership: the owner writes, closes, and returns a receive-only channel. |

## Sources

- [A Tour of Go — Channels](https://go.dev/tour/concurrency/2)
- [Go by Example — Channels](https://gobyexample.com/channels)
- [Go 101 — Channels](https://go101.org/article/channel.html)
- [Boot.dev — Go Channels](https://www.boot.dev/blog/golang/go-channels)
- [Go Channels: A Runtime Internals Deep Dive](https://blog.gaborkoos.com/posts/2025-08-26-Go-Channels-A-Runtime-Internals-Deep-Dive/)
- [Unpacking Go Channels: A Peek Under the Hood](https://dev.to/shrsv/unpacking-go-channels-a-peek-under-the-hood-2c42)
- [Golang Channel Deep Dive: From Zero to Hero](https://dev.to/leapcell/golang-channel-deep-dive-from-zero-to-hero-4com)
- [GeeksforGeeks — Channel in Golang](https://www.geeksforgeeks.org/go-language/channel-in-golang/)
- [func25 — Go Channels](https://func25.dev/posts/go-channels/)
- [Go source — `runtime/chan.go`](https://github.com/golang/go/blob/master/src/runtime/chan.go)
- [YouTube — Go channels introduction](https://www.youtube.com/watch?v=cmlKLz5S6G4)
- [YouTube — Channel internals](https://www.youtube.com/watch?v=r2yJm8fRyRk&t=17)
- [YouTube — Channel behaviour](https://www.youtube.com/watch?v=-BWCKoseOnE)
- [YouTube — Channels and `select`](https://www.youtube.com/watch?v=nDVN5gsLjoo&t=751)
