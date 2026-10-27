# sync.Pool Buffer Example

This example demonstrates how to use `sync.Pool` to reuse byte buffers across a large number of goroutines, reducing allocations and improving performance.

## How It Works

1. **Pool Creation**: A `sync.Pool` is initialized with a `New` function that allocates a `*[]byte` of `bufferSize` (1024 bytes) and increments an allocation counter.
2. **Seeding**: Four buffers are pre-allocated and placed into the pool to avoid allocations for the first few goroutines.
3. **Concurrent Usage**: `numWorkers` (1,024,576) goroutines are spawned using `sync.WaitGroup.Go()` (Go 1.25+). Each goroutine:
   - Retrieves a buffer via `pool.Get()` (type asserting to `*[]byte`)
   - Uses `defer` to return the buffer to the pool with `pool.Put()` before exiting
   - Performs placeholder work
4. **Results**: After all goroutines complete, the program prints how many buffers were actually created. Due to pooling, this will be far less than `numWorkers`.

## Important Notes

- **Return pooled objects**: Always put objects back into the pool when you're done with them. Failing to do so prevents reuse and negates the benefits of pooling.
- **Don't share pooled objects across goroutines**: A buffer checked out by one goroutine must not be used by another until it has been returned to the pool.
- **Pools are not persistent**: The runtime may garbage collect pooled objects, especially during GC cycles. Your `New` function must always be able to create objects from scratch.
- **Type assertions can panic**: `Get()` returns `any`, so type asserting without checking (`.(*[]byte)`) will panic if the pool contains an unexpected type. In this controlled example it's safe since only `*[]byte` values are stored.
- **Counters aren't thread-safe**: `numBuffersCreated` is incremented without synchronization. This is acceptable for observation in a toy example, but use `sync/atomic` if accuracy under races matters.
- **Holding objects indefinitely is harmful**: Don't retain references to pooled objects beyond their intended use. Keeping them alive can increase memory pressure.
- **Use modern APIs when available**: `wg.Go()` (Go 1.25+) simplifies goroutine spawning by handling `Add(1)` and `Done()` automatically.
