package main

import (
	"fmt"
	"sync"
)

// This example demonstrates using sync.Cond with Broadcast() to implement
// a simple event/pub-sub pattern. All registered listeners are woken up
// simultaneously when the event occurs.
//
// Note: sync.Cond is better thought of as a "waiting room"/notification
// mechanism, not as the condition itself. The actual event here is the
// button click.
func main() {
	// Button represents a UI button that can be clicked.
	// Clicked is a condition variable used to notify all subscribers
	// when a click occurs.
	type Button struct {
		Clicked *sync.Cond
	}

	// Create a button with a condition variable backed by a mutex.
	// The mutex is required by sync.Cond and is used to protect shared state
	// and to coordinate waiting/notification.
	button := Button{Clicked: sync.NewCond(&sync.Mutex{})}

	// subscribe registers a callback to be invoked when the button is clicked.
	// It spawns a goroutine that waits on the condition variable until notified.
	subscribe := func(c *sync.Cond, fn func()) {
		var goroutineRunning sync.WaitGroup
		// Ensure the subscriber goroutine is about to call Wait() before returning
		goroutineRunning.Add(1)

		go func() {
			// Lock the condition's mutex before calling Wait().
			// Wait() must be called while holding the lock; it atomically
			// unlocks the mutex and blocks, then reacquires it before returning.
			c.L.Lock()
			defer c.L.Unlock()

			// Signal readiness after acquiring the lock and before waiting.
			// This ensures the goroutine is actually parked on Wait() before
			// the caller proceeds to Broadcast().
			goroutineRunning.Done()

			// Block until Broadcast() or Signal() is called.
			c.Wait()

			// Execute the callback in response to the click event.
			fn()
		}()

		// Wait until the goroutine has acquired the lock and is about to Wait()
		// to avoid missing the broadcast if it happens too early.
		goroutineRunning.Wait()
	}

	// Track when all click handlers have completed
	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)

	// Register three different subscribers to the button click event
	subscribe(button.Clicked, func() {
		fmt.Println("Maximizing window.")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Displaying annoying dialogue box!")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Mouse clicked.")
		clickRegistered.Done()
	})

	// Fire the click event: wake ALL waiting subscribers simultaneously.
	// If we used Signal() instead, only one subscriber would be woken.
	button.Clicked.Broadcast()

	// Wait for all subscribers to finish processing the event
	clickRegistered.Wait()
}
