package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
)

////////////// Custom Error //////////////////////

type MyError struct {
	Inner      error // keep the inner error
	Message    string
	StackTrace string
	Misc       map[string]any // arbitrary structured data for debugging
}

// Error returns the error message as a string.
// MyError is an error.
func (err MyError) Error() string {
	return err.Message
}

func wrapError(err error, messagef string, msgArgs ...any) MyError {
	return MyError{
		Inner:      err,
		Message:    fmt.Sprintf(messagef, msgArgs...),
		StackTrace: string(debug.Stack()),
		Misc:       make(map[string]any),
	}
}

////////////// "job" module /////////////

// JobError - Implicit Interface Satisfaction: Because it embeds the
// built-in error interface, JobError automatically inherits the Error()
// string method. It acts as a valid Go error right out of the box without
// requiring you to write custom boilerplate.
// https://gobyexample.com/struct-embedding
type JobError struct {
	error
}

// runJob is the "job" module. It wraps low-level errors in an
// JobError so that main can recognize them as job-layer errors
// and decide how to present them.
// Returns error chain JobError -> FSError -> MyError -> PathError.
func runJob(id string) error {
	const jobBinPath = "/bad/job/binary"
	isExecutable, err := isGloballyExec(jobBinPath)
	if err != nil {
		// err is a FSError -> MyError -> PathError; promote it to a JobError
		return JobError{err}
	} else if !isExecutable {
		return JobError{wrapError(nil, "job binary is not executable")}
	}

	if err := exec.Command(jobBinPath, "--id="+id).Run(); err != nil {
		return JobError{err}
	}

	return nil
}

////////////// "filesystem" module /////////////////

type FSError struct {
	error // keeps the inner error
}

// returns FSError on failure, nil on success
func isGloballyExec(path string) (bool, error) {
	// err is a PathError
	info, err := os.Stat(path)
	if err != nil {
		return false, FSError{wrapError(err, "%s", err.Error())} // FSError -> MyError -> PathError
	}
	return info.Mode().Perm()&0o100 == 0o100, nil
}

////////////// Error Handling ////////////////////

func handleError(key int, err error, message string) {
	fmt.Printf("\n[logID: %d] ❌  Error\n", key)
	fmt.Printf("%s\n", strings.Repeat("─", 46))

	// Collect each layer so we can align the messages vertically.
	type layer struct {
		typ   string
		msg   string
		stack string
	}
	var layers []layer
	for err != nil {
		l := layer{typ: fmt.Sprintf("[%T]", err), msg: err.Error()}
		if myErr, ok := err.(MyError); ok && myErr.StackTrace != "" {
			l.stack = indentStack(myErr.StackTrace)
		}
		layers = append(layers, l)
		err = unwrapNext(err)
	}

	// Pad type labels so every message starts at the same column.
	maxTyp := 0
	for _, l := range layers {
		if w := len(l.typ); w > maxTyp {
			maxTyp = w
		}
	}
	for _, l := range layers {
		fmt.Printf("  ▸ %-*s %s\n", maxTyp, l.typ, l.msg)
		if l.stack != "" {
			fmt.Printf("     stack:\n%s\n", l.stack)
		}
	}

	fmt.Printf("%s\n", strings.Repeat("─", 46))
	fmt.Printf("  ▸ %s\n", message)
}

// unwrapNext follows one link in the error chain so every wrapping layer
// (MyError, FSError, JobError) is displayed in order.
func unwrapNext(err error) error {
	switch e := err.(type) {
	case MyError:
		return e.Inner
	case FSError:
		return e.error
	case JobError:
		return e.error
	default:
		return nil
	}
}

func indentStack(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = "     " + line
	}
	return strings.Join(lines, "\n")
}

////////////// Main //////////////////////////////

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	err := runJob("1")
	// err is always a JobError from the job module
	if err != nil {
		msg := "There was an unexpected issue; please report this as a bug."
		if _, ok := err.(JobError); ok {
			msg = err.Error() // it's a JobError from the job module
		}
		handleError(1, err, msg)
	}
}
