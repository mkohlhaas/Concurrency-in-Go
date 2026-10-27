# sync.Pool Basic Behavior

This example demonstrates how Go's `sync.Pool` works: objects are only reused if you explicitly put them back, and `sync.Pool.New()` is only called when the pool is empty.

## Code Walkthrough

```go
myPool := &sync.Pool{
    New: func() any {
        fmt.Println("Creating new instance.")
        return struct{}{}
    },
}

myPool.Get()             // will never be put back in the pool
instance := myPool.Get() // creates another instance
myPool.Put(instance)     // and puts it back
myPool.Get()             // will also be lost for good
```

## Execution Flow

1. **`myPool.Get()`** (first call) - Pool is empty, so `New()` is called. `"Creating new instance."` is printed. The returned instance is discarded (never stored or put back), so it's lost to the pool.

2. **`instance := myPool.Get()`** (second call) - Pool is empty again, so `New()` is called a second time. `"Creating new instance."` is printed again. This instance is saved to `instance`.

3. **`myPool.Put(instance)`** - The instance is explicitly returned to the pool, making it available for future `Get()` calls.

4. **`myPool.Get()`** (third call) - The pool now has the instance we just put back, so it is retrieved. `New()` is *not* called (no third print). This instance is not `Put()` back afterwards, so it's lost to the pool.

## Expected Output

```
Creating new instance.
Creating new instance.
```

## Key Takeaways

- **Lazy creation**: `sync.Pool.New` only runs when `Get()` finds no available objects.
- **Explicit reuse**: You must call `Put()` to return objects to the pool. Just calling `Get()` does not recycle them.
- **No automatic retention**: Any object retrieved via `Get()` that is never `Put()` is gone and cannot be reused by the pool.
- **Pool is not a cache**: `sync.Pool` is designed for temporary object reuse to reduce allocations, not for long-term storage.
