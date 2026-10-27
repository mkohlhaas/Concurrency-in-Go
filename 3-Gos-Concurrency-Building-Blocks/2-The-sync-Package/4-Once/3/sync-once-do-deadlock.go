package main

import (
	"sync"
)

// This example demonstrates a deadlock caused by a circular dependency
// between two sync.Once instances. sync.Once.Do() holds an internal lock
// while executing the initialization function, so re-entering Do() on the
// same Once instance from within its own initialization will block forever.
func main() {
	var onceA, onceB sync.Once

	// Forward declaration of initB so initA can reference it before definition
	var initB func()

	// initA will trigger initialization of onceB via initB
	initA := func() { onceB.Do(initB) }

	// initB will trigger initialization of onceA via initA, creating a cycle
	initB = func() { onceA.Do(initA) }

	// Start the chain by initializing onceA.
	// Execution flow leading to deadlock:
	// 1. onceA.Do(initA) acquires onceA's lock and runs initA
	// 2. initA calls onceB.Do(initB); onceB acquires its lock and runs initB
	// 3. initB calls onceA.Do(initA); onceA is already locked/in-progress,
	//    so this blocks waiting for the original initA to complete
	// 4. initA is still blocked waiting for onceB.Do(initB) to return
	//    (initB hasn't finished) → circular wait → deadlock
	onceA.Do(initA)
}
