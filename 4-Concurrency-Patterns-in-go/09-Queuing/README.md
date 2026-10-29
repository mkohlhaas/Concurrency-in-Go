## 5. Bottom-line assessment

**Buffered write wins decisively for small, frequent writes** — roughly 3× faster here with no extra allocation cost relative to the unbuffered version (both allocate the same 1 byte/op). If your real code writes many small chunks to a file, wrap it in `bufio.Writer`. If you already write large buffers, the benefit will be smaller and you should benchmark that actual pattern rather than this synthetic one.
