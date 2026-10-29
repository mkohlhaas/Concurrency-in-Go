package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	done := make(chan struct{})

	// Cancel after 2 seconds — before the 5-second locale lookups finish.
	// Comment out to not cancel.
	go func() {
		time.Sleep(2 * time.Second)
		close(done)
	}()

	// Greeting //
	wg.Go(func() {
		if err := printGreeting(done); err != nil {
			fmt.Printf("greeting: %v\n", err)
			return
		}
	})

	// Farewell //
	wg.Go(func() {
		if err := printFarewell(done); err != nil {
			fmt.Printf("farewell: %v\n", err)
			return
		}
	})

	wg.Wait()
	fmt.Println("both tasks finished")
}

// -------- //
// Greeting //
// -------- //

func printGreeting(done <-chan struct{}) error {
	greeting, err := genGreeting(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func genGreeting(done <-chan struct{}) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

// -------- //
// Farewell //
// -------- //

func printFarewell(done <-chan struct{}) error {
	farewell, err := genFarewell(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genFarewell(done <-chan struct{}) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

// ------ //
// Locale //
// ------ //

// can be cancelled via done channel
func locale(done <-chan struct{}) (string, error) {
	select {
	case <-done:
		return "", fmt.Errorf("canceled")
	case <-time.After(5 * time.Second):
	}
	return "EN/US", nil
}
