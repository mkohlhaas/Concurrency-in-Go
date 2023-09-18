package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Calling printGreeting")
		if err := printGreeting(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			fmt.Println("Calling cancel explicitly from the top.")
			cancel() // will cancel all of main() and its sub-goroutines
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Calling printFarewell")
		if err := printFarewell(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()

	wg.Wait()
}

func printGreeting(ctx context.Context) error {
	fmt.Println("  Calling genGreeting")
	greeting, err := genGreeting(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func printFarewell(ctx context.Context) error {
	fmt.Println("  Calling genFarewell")
	farewell, err := genFarewell(ctx)
	if err != nil {
		fmt.Println("Error in printFarewell")
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genGreeting(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second) // this will cancel in locale() -> ctx.Done()
	defer cancel()

	fmt.Println("    Calling locale in genGreeting")
	switch locale, err := locale(ctx); {
	case err != nil:
		fmt.Println("Error in genGreeting")
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func genFarewell(ctx context.Context) (string, error) {
	fmt.Println("    Calling locale in genFarewell")
	switch locale, err := locale(ctx); {
	case err != nil:
		fmt.Println("Error in genFarewell")
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(ctx context.Context) (string, error) {
	fmt.Println("      locale got called")
	select {
	case <-ctx.Done(): // catching cancel() in genGreeting or cancel() called from the top
		fmt.Println("\nctx.Done in locale")
		return "", ctx.Err()
	case <-time.After(60 * time.Second): // set timeout to one second to see a difference!!!
		fmt.Println("timed out in locale")
	}
	return "EN/US", nil
}
