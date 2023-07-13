package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
)

////////////// Custom Error //////////////////////

type MyError struct {
	Inner      error
	Message    string
	StackTrace string
	Misc       map[string]any
}

func wrapError(err error, messagef string, msgArgs ...any) MyError {
	return MyError{
		Inner:      err,
		Message:    fmt.Sprintf(messagef, msgArgs...),
		StackTrace: string(debug.Stack()),
		Misc:       make(map[string]any),
	}
}

// this makes MyError an error
func (err MyError) Error() string {
	return err.Message
}

////////////// "intermediate" module /////////////

type IntermediateErr struct {
	error
}

func runJob(id string) error {
	const jobBinPath = "/bad/job/binary"
	isExecutable, err := isGloballyExec(jobBinPath)
	if err != nil {
		return err
	} else if isExecutable == false {
		return wrapError(nil, "job binary is not executable")
	}

	return exec.Command(jobBinPath, "--id="+id).Run()
}

////////////// "lowlevel" module /////////////////

type LowLevelErr struct {
	error
}

func isGloballyExec(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, LowLevelErr{(wrapError(err, err.Error()))}
	}
	return info.Mode().Perm()&0100 == 0100, nil
}

////////////// Error Handling ////////////////////

func handleError(key int, err error, message string) {
	log.SetPrefix(fmt.Sprintf("[logID: %v]: ", key))
	log.Printf("%#v", err)
	log.Printf("%v", err)
	fmt.Printf("[%v] %v", key, message)
}

////////////// Main //////////////////////////////

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	err := runJob("1")
	if err != nil {
		msg := "There was an unexpected issue; please report this as a bug.\n"
		if _, ok := err.(IntermediateErr); ok {
			msg = err.Error()
		}
		handleError(1, err, msg)
	}
}
