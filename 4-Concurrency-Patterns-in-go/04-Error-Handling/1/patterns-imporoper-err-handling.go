package main

import (
	"fmt"
	"net/http"
)

func main() {
	checkStatus := func(done <-chan any, urls ...string) <-chan *http.Response {
		responses := make(chan *http.Response)

		go func() {
			defer close(responses)
			for _, url := range urls {
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(err)
					continue
				}
				select {
				case <-done:
					return
				case responses <- resp:
				}
			}
		}()

		return responses
	}

	done := make(chan any)
	defer close(done)

	urls := []string{"https://www.google.com", "https://badhost"}
	// urls := []string{"https://www.google.com", "https://www.bawue.de"}
	for response := range checkStatus(done, urls...) {
		fmt.Printf("Response: %v\n", response.Status)
	}
}
