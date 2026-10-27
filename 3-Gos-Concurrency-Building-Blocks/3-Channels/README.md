## What's a Channel in Go?

A channel in the Go programming language is a typed conduit that lets concurrent goroutines (lightweight threads) send and receive values to communicate and synchronize with each other. [1] (https://go.dev/tour/concurrency/2), [2] (https://gobyexample.com/channels), [3] (https://www.youtube.com/watch?v=cmlKLz5S6G4&t=1)
Instead of sharing memory by locking variables, Go follows the Communicating Sequential Processes (CSP) model: "Do not communicate by sharing memory; instead, share memory by communicating." [1] (https://blog.gaborkoos.com/posts/2025-08-26-Go-Channels-A-Runtime-Internals-Deep-Dive/)

Core Characteristics

• Typed: Each channel can only transmit data of a specific type (e.g., chan int or chan string). [1] (https://www.geeksforgeeks.org/go-language/channel-in-golang/), [2] (https://www.youtube.com/watch?v=-BWCKoseOnE)
• First-In-First-Out (FIFO): Values are received in the exact order they are sent. [1] (https://www.boot.dev/blog/golang/go-channels), [2] (https://dev.to/leapcell/golang-channel-deep-dive-from-zero-to-hero-4com)
• Blocking by Default: By default, sends and receives block execution until the other side is ready, which naturally synchronizes goroutines without explicit locks. [1] (https://gobyexample.com/channels), [2] (https://go.dev/tour/concurrency/2)
• Reference Type: Like slices and maps, channels are reference types and must be initialized with make() before use (an uninitialized channel is nil). [1] (https://go101.org/article/channel.html), [2] (https://go.dev/tour/concurrency/2), [3] (https://www.boot.dev/blog/golang/go-channels)

Basic Syntax & Operations

You use the channel operator <- to send and receive data: [1] (https://go.dev/tour/concurrency/2), [2] (https://www.boot.dev/blog/golang/go-channels)
• Creation:go
ch := make(chan int) // Unbuffered channel
Use code with caution.
• Send to a channel:go
ch <- 42 // Sends the value 42 into channel ch
Use code with caution.
• Receive from a channel:go
val := <-ch // Receives a value from ch and assigns it to val
Use code with caution.

Types of Channels

1. Unbuffered Channels:
	• Created with make(chan T).
	• The sender blocks immediately until a receiver is ready to take the value. [1] (https://dev.to/shrsv/unpacking-go-channels-a-peek-under-the-hood-2c42), [2] (https://www.youtube.com/watch?v=r2yJm8fRyRk&t=17), [3] (https://go.dev/tour/concurrency/2)
2. Buffered Channels:
	• Created with make(chan T, capacity) (e.g., make(chan int, 3)).
	• Allows the sender to push up to a specific number of items before blocking, as long as the buffer isn't full. [1] (https://dev.to/shrsv/unpacking-go-channels-a-peek-under-the-hood-2c42), [2] (https://blog.gaborkoos.com/posts/2025-08-26-Go-Channels-A-Runtime-Internals-Deep-Dive/)
3. Directional Channels:
	• You can restrict channels in function signatures to be send-only (chan<- T) or receive-only (<-chan T) for safety. [1] (https://go101.org/article/channel.html)
If you'd like, I can show you a code example demonstrating how to use channels with goroutines, or explain how select statements work with multiple channels.

