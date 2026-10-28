package main

import (
	"context"
	"fmt"
	"net/http"
)

// checkStatus takes a context and a list of URLs.
// It returns a channel that receives the *http.Response for each URL
// that was successfully fetched, in completion order.
//
// The context controls cancellation: once cancelled, the worker
// stops making new requests and exits.
func checkStatus(ctx context.Context, urls ...string) <-chan *http.Response {
	// Unbuffered channel: each send blocks until the main loop receives it.
	responses := make(chan *http.Response)

	// Worker goroutine: fans out HTTP GET requests for each URL.
	go func() {
		// Close the responses channel when done so the range loop terminates.
		defer close(responses)

		for _, url := range urls {
			// Check for cancellation before each request.
			select {
			case <-ctx.Done():
				return
			default:
			}

			resp, err := http.Get(url)
			// Error handling: a failed request (e.g. bad host) is logged
			// and skipped — we don't crash, just move on to the next URL.
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Hand off the response, aborting early if cancelled.
			select {
			case <-ctx.Done():
				return
			case responses <- resp:
				// Response handed off; fetch the next URL.
			}
		}
	}()

	return responses
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	urls := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://httpbin.org/delay/3", // slow endpoint, would be skipped on cancel
		"https://www.example.com",
		"https://www.github.com",
		"https://www.google.com",
	}

	responses := checkStatus(ctx, urls...)

	// Range over the responses channel; it closes when the worker finishes.
	received := 0
	for response := range responses {
		fmt.Printf("Response: %v\n", response.Status)
		received++

		// Cancel mid-point: after receiving two responses, stop the worker
		// so it no longer fetches the remaining (slower) URLs.
		if received == 2 {
			fmt.Println("Cancelling context — worker will stop fetching remaining URLs.")
			cancel()
		}
	}
}
