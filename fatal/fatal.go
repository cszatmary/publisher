package fatal

import (
	"fmt"
	"os"
)

var showStackTraces = true

func ShowStackTraces(show bool) {
	showStackTraces = show
}

func printErr(err error) {
	if err == nil {
		return
	}

	if showStackTraces {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

func ExitErr(err error, message string) {
	fmt.Fprintln(os.Stderr, message)

	printErr(err)
	os.Exit(1)
}

func ExitErrf(err error, format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)

	printErr(err)
	os.Exit(1)
}

func Exit(message string) {
	ExitErr(nil, message)
}

func Exitf(format string, a ...interface{}) {
	ExitErrf(nil, format, a...)
}
