# The `select` Statement

`select` lets a goroutine wait on multiple channel operations at once. It looks
like a `switch`, but the cases are sends and receives rather than boolean
conditions: the runtime executes the first case whose operation is ready.

## Core behavior and rules

- **Blocking** — if no channel is ready, `select` blocks the whole goroutine
  until at least one case can proceed.
- **Random selection** — if several cases are ready at the same moment, Go picks
  one at random instead of evaluating top to bottom. This prevents channel
  starvation.
- **`default`** — adding a `default` case makes the `select` non-blocking: if no
  channel is ready immediately, the `default` branch runs.

## Basic syntax

A typical structure listens on a data channel alongside a termination or timeout
signal:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	quit := make(chan bool)

	// Simulate a concurrent task.
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(500 * time.Millisecond)
			ch <- fmt.Sprintf("Message %d", i)
		}
		quit <- true
	}()

	// Monitor both channels simultaneously.
	for {
		select {
		case msg := <-ch:
			fmt.Println("Received:", msg)
		case <-quit:
			fmt.Println("Quit signal received. Exiting.")
			return
		}
	}
}
```

## Common patterns

### 1. Non-blocking channel operations (`default`)

`default` attempts a send or receive once and moves on if it cannot happen
immediately:

```go
select {
case msg := <-ch:
	fmt.Println("Received:", msg)
default:
	fmt.Println("No data available right now, doing other work...")
}
```

### 2. Timeouts

Pairing a data channel with `time.After` keeps an operation from stalling
forever:

```go
select {
case res := <-apiChannel:
	fmt.Println("API Response:", res)
case <-time.After(2 * time.Second):
	fmt.Println("Error: API request timed out!")
}
```

### 3. Blocking forever (`select {}`)

An empty `select` with no cases blocks the current goroutine forever. It is
commonly used at the end of `main` when permanent background goroutines are
running and the application must not exit.

## Pitfalls

- **`break` inside a loop** — when a `select` sits inside a `for` loop, `break`
  in a case exits only the `select`, not the loop. Use a label to leave the loop
  entirely:

  ```go
  BackgroundLoop:
  for {
      select {
      case <-quit:
          break BackgroundLoop // exits the for loop
      }
  }
  ```

- **Nil channels** — sending to or receiving from a `nil` channel blocks forever.
  A `nil` channel used in a `select` simply disables that case permanently.

- **Closed channels** — receiving from a closed channel returns the zero value
  immediately. Without the comma-ok form (`v, ok := <-ch`) the case fires every
  time and the `select` spins in a busy, high-CPU loop.

## Examples in this directory

Each numbered directory is a standalone module; run one with `cd <n> && go run .`.

| Example | File | Demonstrates |
| --- | --- | --- |
| 1 | [`select-blocking.go`](./select-blocking.go) | A bare `select` blocking until a goroutine closes the channel. |

## Sources

- [A Tour of Go — Select](https://go.dev/tour/concurrency/5)
- [GeeksforGeeks — Select Statement in Go](https://www.geeksforgeeks.org/go-language/select-statement-in-go-language/)
- [Go's select statement is a switch for channels](https://medium.com/golang-in-plain-english/gos-select-statement-is-a-switch-for-channels-and-it-s-more-powerful-than-you-think-78930d3fe009)
- [Understanding the select statement in Go](https://leapcell.io/blog/understanding-the-select-statement-in-go)
- [The secret life of Go: the select statement](https://dev.to/aaron_rose_0787cc8b4775a0/the-secret-life-of-go-the-select-statement-27ob)
- [Go runtime internals: select](https://internals-for-interns.com/posts/go-runtime-select/)
- [YouTube — Go `select` introduction](https://www.youtube.com/watch?v=mcl3zU5wat4&t=170)
