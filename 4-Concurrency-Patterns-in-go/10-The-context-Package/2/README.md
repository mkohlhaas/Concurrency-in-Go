### The Context Package

- It's all about cancellations and storing context specific values.
- See Youtube video: ["Go context.Context" – the Go context package: What it is, implementation and use](https://www.youtube.com/watch?v=PWBVTJyzvXs&t=1s&pp=ygUTZGFzIGNvbnRleHQgcGFja2FnZQ%3D%3D)
- See also Lightstep videos [Context Propagation makes OpenTelemetry awesome](https://www.youtube.com/watch?v=gviWKCXwyvY) and  [The painful simplicity of context propagation in Go](https://www.youtube.com/watch?v=g4ShnfmHTs4)
- Contexts can be enriched using **With...** functions, e.g. **context.WithCancel**, **context.WithTimeout**, **context.WithValue**, ...
  - **With...** functions take an existing context and return an enriched context.
  - Contexts are per se immutable.
- You start with an empty context - **context.Background()** - and then enrich with **With...** functions.
- **context.TODO()** is also an empty context but is used when you don't know yet what context to use.
- The context is the first parameter of the functions using a context and typically called **ctx**.
- The created contexts form a tree:
  - **Cancellations** are propagated **DOWN** the tree.
  - **Values** are looked up - via **ctx.Value(...)** - going **UP** the tree.
- With... functions:
  | Function          | Creates a context ...                                                                            |
  | ------------------| ------------------------------------------------------------------------------------------------ |
  | WithValue         | with a (key, value) pair for storing context relevant information, e.g. transaction ID.          |
  | WithCancel        | the can be cancelled. Cancellation is advertised by closing the channel returned by ctx.Done().  |
  | WithTimeout       | that will be automatically cancelled after a certain timeout.                                    |
  | WithDeadline      | that will be automatically cancelled at a certain point in time.                                 |
  | WithoutCancel     | that removes the Done channel from the context.                                                  |
  | WithCancelCause   | with a custom cancel cause error.                                                                |
  | WithTimeoutCause  | with a custom timeout cause error.                                                               |
  | WithDeadlineCause | with a custom deadline cause error.                                                              |
