# The Context Package

- It's all about cancellations and storing context-specific values.
- See YouTube video: ["Go context.Context" – the Go context package: What it is, implementation and use](https://www.youtube.com/watch?v=PWBVTJyzvXs&t=1s&pp=ygUTZGFzIGNvbnRleHQgcGFja2FnZQ%3D%3D)
- See also Lightstep videos: [Context Propagation makes OpenTelemetry awesome](https://www.youtube.com/watch?v=gviWKCXwyvY) and [The painful simplicity of context propagation in Go](https://www.youtube.com/watch?v=g4ShnfmHTs4)
- Contexts can be enriched using **With...** functions, e.g. `context.WithCancel`, `context.WithTimeout`, `context.WithValue`, ...
  - **With...** functions take an existing context and return an enriched context.
  - Contexts are per se immutable.
- You start with an empty context — `context.Background()` — and then enrich with **With...** functions.
- `context.TODO()` is also an empty context but is used when you don't know yet what context to use.
- The context is the first parameter of the functions using a context and is typically called **ctx**.
- The created contexts form a tree:
  - **Cancellations** are propagated **DOWN** the tree.
  - **Values** are looked up — via `ctx.Value(...)` — going **UP** the tree.

## With... functions

| Function | Creates a context ... |
| --- | --- |
| `WithValue` | with a `(key, value)` pair for storing context-relevant information, e.g. transaction ID. |
| `WithCancel` | that can be cancelled. Cancellation is advertised by closing the channel returned by `ctx.Done()`. |
| `WithTimeout` | that will be automatically cancelled after a certain timeout. |
| `WithDeadline` | that will be automatically cancelled at a certain point in time. |
| `WithoutCancel` | that removes the `Done` channel from the context. |
| `WithCancelCause` | with a custom cancel cause error. |
| `WithTimeoutCause` | with a custom timeout cause error. |
| `WithDeadlineCause` | with a custom deadline cause error. |

## What is Context?

In Go, `context.Context` is a standard library interface used to manage deadlines, cancellation signals, and request-scoped data across API boundaries and goroutines. Think of it as a supervisor that travels alongside your functions and goroutines, carrying crucial metadata and telling them when it's time to stop working. [1](https://www.reddit.com/r/golang/comments/18mphqt/what_even_is_context/), [2](https://medium.com/@adityarpadekar/understanding-context-in-go-a-comprehensive-guide-c2deb6863bc7)

Without context, a slow or hung HTTP request could leave orphan goroutines running forever, leading to resource leaks. [1](https://www.youtube.com/watch?v=BkzgYfygDy8&t=370), [2](https://www.youtube.com/watch?v=fCYZprTvNVg)

## The Core Interface

The context interface itself is remarkably simple, consisting of only four methods: [1](https://dzone.com/articles/contexts-in-go-a-comprehensive-guide), [2](https://medium.com/@adityarpadekar/understanding-context-in-go-a-comprehensive-guide-c2deb6863bc7)

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool) // When this context will be canceled
    Done() <-chan struct{}                   // Channel that closes when work should stop
    Err() error                              // Explains WHY the context was canceled
    Value(key any) any                       // Retrieves request-scoped data
}
```

## The Two Base Contexts

Every context tree requires a root node. You generate an empty, un-cancelable context using one of two functions: [1](https://medium.com/@adityarpadekar/understanding-context-in-go-a-comprehensive-guide-c2deb6863bc7)

- **`context.Background()`**: The standard root context. Use this in main functions, HTTP request entry points, initialization logic, and tests.
- **`context.TODO()`**: A placeholder context. Use this when you aren't sure which context to use yet, signaling that the code needs further refactoring. [1](https://www.thegeekyway.com/understanding-context-in-golang/), [2](https://medium.com/@adityarpadekar/understanding-context-in-go-a-comprehensive-guide-c2deb6863bc7)

## The Derived Context Types

To give a context powers (like timeouts or cancel signals), you pass a parent context into a constructor function. This creates a "child" context in a tree-like hierarchy. [1](https://dev.to/jones_charles_ad50858dbc0/mastering-gos-context-package-a-practical-guide-for-everyday-devs-1ll5), [2](https://medium.com/@adityarpadekar/understanding-context-in-go-a-comprehensive-guide-c2deb6863bc7)

### Manual Cancellation (context.WithCancel)

Allows a parent routine to explicitly signal to a child routine that it should stop working. [1](https://www.youtube.com/watch?v=TBZ-m9h4Ymo&t=36), [2](https://medium.com/@adityarpadekar/understanding-context-in-go-a-comprehensive-guide-c2deb6863bc7)

```go
ctx, cancel := context.WithCancel(context.Background())
// Call cancel() when something goes wrong or work finishes early
defer cancel()
```

### Relative Timeout (context.WithTimeout)

Automatically cancels the context after a specific duration passes. Extremely common for database queries and external HTTP requests. [1](https://www.reddit.com/r/golang/comments/1fk578x/golang_context_package_explained_and_how_to_use/), [2](https://www.thegeekyway.com/understanding-context-in-golang/), [3](https://www.youtube.com/watch?v=BkzgYfygDy8&t=370)

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel() // Releases resources if the work finishes before 2 seconds
```

### Absolute Deadline (context.WithDeadline)

Automatically cancels the context at a specific target clock time. [1](https://www.reddit.com/r/golang/comments/1fk578x/golang_context_package_explained_and_how_to_use/), [2](https://www.thegeekyway.com/understanding-context-in-golang/)

```go
ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
defer cancel()
```

### Scope Values (context.WithValue)

Carries tiny, request-scoped metadata through the execution chain. It should never be used to pass optional function parameters. Use it strictly for operational data like request IDs, tracing tokens, or authentication details. [1](https://www.youtube.com/watch?v=h2RdcrMLQAo&t=858), [2](https://www.reddit.com/r/golang/comments/18mphqt/what_even_is_context/), [3](https://medium.com/@adityarpadekar/understanding-context-in-go-a-comprehensive-guide-c2deb6863bc7), [4](https://pkg.go.dev/context), [5](https://www.reddit.com/r/golang/comments/1fk578x/golang_context_package_explained_and_how_to_use/), [6](https://dev.to/jones_charles_ad50858dbc0/mastering-gos-context-package-a-practical-guide-for-everyday-devs-1ll5)

```go
ctx := context.WithValue(context.Background(), "traceID", "abc-123-xyz")
```

## Code Example: Putting it to Work

Here is how you use a timeout context to stop a slow-running background process: [1](https://www.youtube.com/watch?v=BkzgYfygDy8&t=370), [2](https://www.thegeekyway.com/understanding-context-in-golang/)

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// 1. Create a context that times out after 2 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // Good practice to prevent context leaks

	// 2. Run an operation that takes 5 seconds
	go doHeavyWork(ctx)

	// Wait long enough to observe the outcome
	time.Sleep(3 * time.Second)
}

func doHeavyWork(ctx context.Context) {
	select {
	case <-time.After(5 * time.Second):
		fmt.Println("Work finished successfully!")
	case <-ctx.Done():
		// 3. The 2-second timeout triggers first, closing the Done channel
		fmt.Println("Work aborted:", ctx.Err()) // Prints: Work aborted: context deadline exceeded
	}
}
```

## Golden Rules of Go Context

- **Parameter placement**: Always pass `ctx context.Context` as the very first parameter of a function. Do not store it inside a struct unless implementing exceptional library patterns. [1](https://go.dev/blog/context-and-structs), [2](https://pkg.go.dev/context), [3](https://www.youtube.com/watch?v=h2RdcrMLQAo&t=858)
- **Don't pass nil**: If a function requires a context but you aren't sure what to provide yet, pass `context.TODO()` instead of nil. [1](https://pkg.go.dev/context)
- **Always call cancel**: If you use `WithCancel`, `WithTimeout`, or `WithDeadline`, always execute the returned `cancel()` function (usually via `defer cancel()`) to release underlying resources promptly. [1](https://pkg.go.dev/context), [2](https://www.thegeekyway.com/understanding-context-in-golang/)
- **Thread Safety**: Contexts are completely safe to pass across multiple simultaneous goroutines. [1](https://pkg.go.dev/context)
