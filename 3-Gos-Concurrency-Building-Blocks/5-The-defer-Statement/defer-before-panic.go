package main

import (
	"fmt"
)

// main demonstrates that deferred function calls still run when the
// surrounding function panics. When a panic occurs, Go unwinds the stack of
// the panicking goroutine, running every deferred call in LIFO order before
// terminating the program (unless the panic is recovered).
func main() {
	// Defer an anonymous function so its execution is postponed until main
	// returns or, as in this case, until the stack begins unwinding because
	// of a panic. This is the "before the panic" message that still gets
	// printed even though the program is about to crash.
	defer func() { fmt.Println("before the panic") }()

	// Trigger a panic. Everything after this line in main is not executed;
	// instead, control jumps to the deferred functions registered above.
	panic("panicked")
}
