# Confinement

Confinement is a design principle for safe concurrent code: instead of guarding
shared data with locks, make sure the data can only be reached from one
goroutine at a time. If only one goroutine can touch a value, there is no data
race by construction — no `sync.Mutex` is needed.

This is the same idea behind:

> Do not communicate by sharing memory; instead, share memory by communicating.

Channels are one way to enforce it (only the owner sends and closes), but
confinement also works without channels, by limiting which goroutine may read
or write a value.

The book *Concurrency in Go* (Katherine Cox-Buday) describes two main kinds:

1. **Ad-hoc confinement** — the weakest kind. Nothing in the language or type
   system enforces it; you and your team simply agree to follow a rule such as
   "only goroutine X touches this variable." The compiler will not stop you if
   the rule is broken.
2. **Lexical confinement** — stronger. The language's scoping rules make it
   impossible for other goroutines to access the value, e.g. a variable declared
   inside a function and only used there, or passing a slice/copy so that each
   goroutine owns its own data.

## Ad-hoc confinement

In the first example, `data` is shared between `main` and the `loopData`
goroutine:
[`1/confinement-ad-hoc.go`](./1/confinement-ad-hoc.go)

```go
data := make([]int, 4)          // created in main

loopData := func(handleData chan<- int) {
	defer close(handleData)
	for i := range data {
		handleData <- data[i]   // reads data
	}
}

handleData := make(chan int)
go loopData(handleData)         // only this goroutine reads data
```

`data` is a slice — it has a backing array with a pointer, and nothing prevents
another goroutine from indexing or writing `data`. The code is safe only
because of an implicit convention:

- `main` initializes `data` **before** starting the goroutine, then never
  touches it again.
- `loopData` is the only goroutine that reads `data`.

That is ad-hoc confinement: **the safety comes from programmer discipline, not
from the language.** If someone later added
`go func() { data[0] = 42 }()` in `main`, the program would have a data race.

## Ownership and lexical confinement

The sibling examples show stronger forms of the same idea.

**Channel ownership** ([`2/confinement-ownership.go`](./2/confinement-ownership.go)):
the goroutine that creates a channel owns it and is the only one that sends to
it and closes it; consumers only receive. The directional types (`<-chan int`,
`chan<- int`) encode the rule in the function signatures.

```go
chanOwner := func() <-chan int {
	results := make(chan int, 5)
	go func() {
		defer close(results)
		for i := 0; i <= 5; i++ {
			results <- i
		}
	}()
	return results
}

consumer := func(results <-chan int) {
	for result := range results {
		fmt.Printf("Received: %d\n", result)
	}
}

results := chanOwner()
consumer(results)
```

**Lexical confinement** ([`3/confinement-structs.go`](./3/confinement-structs.go)):
each goroutine gets its own slice of the byte array (`data[:3]`, `data[3:]`) and
writes to a local `bytes.Buffer`, so no two goroutines ever write the same
memory. The confinement follows from the scopes involved, not from a
convention people have to remember.

```go
data := []byte("golang")
go printData(&wg, data[:3])
go printData(&wg, data[3:])
```

## Examples in this directory

Each numbered directory is a standalone module; run one with `cd <n> && go run .`.

| Example | File | Demonstrates |
| --- | --- | --- |
| 1 | [`confinement-ad-hoc.go`](./1/confinement-ad-hoc.go) | Ad-hoc confinement: one goroutine reads data, enforced only by convention. |
| 2 | [`confinement-ownership.go`](./2/confinement-ownership.go) | Channel ownership: the creator sends and closes, the consumer only receives. |
| 3 | [`confinement-structs.go`](./3/confinement-structs.go) | Lexical confinement: each goroutine works on its own slice. |

## Rule of thumb

Prefer lexical confinement and channel ownership when you can express them,
because they cannot be accidentally violated. Ad-hoc confinement works, but it
depends on everyone knowing and respecting the convention — and it survives
code review, not compilation.

## Sources

- Katherine Cox-Buday, *Concurrency in Go*, Chapter 4: Concurrency Patterns in Go
- [`1/confinement-ad-hoc.go`](./1/confinement-ad-hoc.go) — the example this
  README walks through
