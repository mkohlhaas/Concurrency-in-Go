# Error Propagation — Embedded `error` Field

## Is `error` a type or the name of a field in JobError?

It is simultaneously both.

In Go, this is called an embedded field (or anonymous field). When you omit an explicit field name, Go treats the unqualified type name as the implicit name of the field. [1](https://go101.org/article/type-embedding.html), [2](https://boldlygo.tech/archive/2023-03-13-embedded-struct-fields/)

Because of this rule, `error` functions as both the type specification and the identifier you use to interact with it: [1](https://go101.org/article/type-embedding.html)

### 1. error is the Type

The struct is reserving an internal memory slot that holds any value satisfying Go's built-in error interface. [1](https://go101.org/article/type-embedding.html)

### 2. error is the Field Name

Because no explicit name was provided (e.g., `Err error`), Go automatically assigns the field name to be the lowercase identifier `error`. [1](https://stackoverflow.com/questions/24537443/meaning-of-a-struct-with-embedded-anonymous-interface), [2](https://go101.org/article/type-embedding.html)

## How this works in practice

Because it is an actual field name, you use `error` explicitly when you instantiate or directly read the field: [1](https://gobyexample.com/struct-embedding), [2](https://go101.org/article/type-embedding.html)

### Instantiating the Struct

You must explicitly use `error` as the key when writing a struct literal: [1](https://gobyexample.com/struct-embedding)

```go
err := JobError{
    error: errors.New("database connection lost"), // 'error' is the field name here
}
```

### Directly Accessing the Field

If you need to bypass Go's method promotion and look directly at the underlying interface value, you use `.error`: [1](https://coddy.tech/docs/go/struct-embedding), [2](https://go101.org/article/type-embedding.html)

```go
// Accessing the inner error field directly
if err.error != nil {
    fmt.Println(err.error.Error())
}
```

## What this example demonstrates

This example shows **error propagation at scale** — how to wrap and carry errors through multiple layers of a program while preserving context, and how to identify specific error categories at the top level.

### 1. Rich custom error type (`MyError`)

Instead of a plain string error, `MyError` captures:

- **`Inner`** — the original cause (for `errors.Is`/`errors.As` chains)
- **`Message`** — a human-readable description added at this layer
- **`StackTrace`** — where the error was created
- **`Misc`** — arbitrary structured data for debugging

`wrapError` builds one of these, and `MyError.Error()` makes it satisfy Go's `error` interface.

### 2. Embedded `error` fields (`JobError`, `FSError`)

```go
type JobError struct {
    error
}
type FSError struct {
    error
}
```

Because `error` is an **embedded (anonymous) field**, each struct is itself an `error` (method promotion), *and* the embedded field is literally named `error`. That means the struct can be returned wherever an `error` is expected without writing a custom `Error()` method.

### 3. Wrapping at each layer

Errors are **wrapped, not replaced**, as they bubble up:

| Layer | What it does |
|-------|--------------|
| Filesystem (`isGloballyExec`) | Wraps `os.Stat` failure in `FSError{wrapError(...)}` |
| Job (`runJob`) | Promotes the low-level error into `JobError{...}` |
| Entry point (`main`) | Type-asserts on `JobError` to decide how to present it |

Each layer adds its own context without losing the original cause.

### 4. Walking the error chain

`handleError` walks the full chain (`JobError → FSError → MyError → PathError`) using `unwrapNext`, printing each layer's type and message, plus the stack trace captured at the `MyError` layer:

```
[logID: 1] ❌  Error
──────────────────────────────────────────────
  ▸ [main.JobError]   stat /bad/job/binary: no such file or directory
  ▸ [main.FSError]    stat /bad/job/binary: no such file or directory
  ▸ [main.MyError]    stat /bad/job/binary: no such file or directory
     stack:
     ...
  ▸ [*fs.PathError]   stat /bad/job/binary: no such file or directory
──────────────────────────────────────────────
  ▸ There was an unexpected issue; please report this as a bug.
```

### 5. Selective handling at the boundary

```go
if _, ok := err.(JobError); ok {
    msg = err.Error() // it's a JobError from the intermediate module
}
```

`main` treats `JobError` as a known, reportable error (uses its message), while anything else gets the generic "please report this as a bug" message.

### The pattern

**Wrap with context at each boundary, inspect by type at the top.** This is the standard Go approach for production error handling: preserve the full chain so you can debug deep failures, but present the user a clean, category-aware message at the edge.
