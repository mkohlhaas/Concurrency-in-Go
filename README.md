This is the full source code for the book [Concurrency in Go](https://katherine.cox-buday.com/concurrency-in-go/) published by O'Reilly.

For errata and more information, please refer to the book's hub at [Errata for Concurrency in Go](https://www.oreilly.com/catalog/errata.csp?isbn=9781491941195).

![Concurrency in Go](https://katherine.cox-buday.com/assets/concurrency-in-go/book-cover.jpg)

## Notes on `init()` in Go

Yes, `init()` is automatically called by the Go runtime.

- **Automatic execution**: `init()` functions have no parameters or return values and cannot be called manually. The runtime executes them during package initialization, before `main()` runs or before any tests/benchmarks are executed.
- **Timing**: Since test files can be in `package main`, `init()` will run when the test binary initializes the package. It can start daemons/goroutines, and synchronization primitives like `sync.WaitGroup` can block package initialization until ready.
- **Order**: If there are multiple `init()` functions across packages/imports, they run in dependency order. Within a package, they run in the order they're declared.
