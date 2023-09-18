package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		// salutation := salutation // you could use shadowing to prevent capturing of loop variable "salutation"

		go func() {
			defer wg.Done()
			fmt.Println(salutation) // "loop variable salutation captured by func literal"
		}()

	}
	wg.Wait()
}
