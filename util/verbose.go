package util

import (
	"fmt"
	"io"
)

// Verbose indicates if currently in verbose mode.
var verbose bool

func SetVerboseMode(enabled bool) {
	verbose = enabled
}

// VerbosePrintln formats using the default formats for
// its operands and writes to standard output if in verbose mode.
func VerbosePrintln(a ...interface{}) (n int, err error) {
	if verbose {
		return fmt.Println(a...)
	}

	return 0, nil
}

// VerboseFprintln formats using the default formats for
// its operands and writes to w if in verbose mode.
func VerboseFprintln(w io.Writer, a ...interface{}) (n int, err error) {
	if verbose {
		return fmt.Fprintln(w, a...)
	}

	return 0, nil
}

// VerbosePrintf formats according to a format specifier and
// writes to standard output if in verbose mode.
func VerbosePrintf(format string, a ...interface{}) (n int, err error) {
	if verbose {
		return fmt.Printf(format, a...)
	}

	return 0, nil
}
