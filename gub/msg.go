package gub

import (
	"fmt"
	"os"
)

func errmsg(format string, a ...interface{}) (n int, err error) {
	format = "** " + format + "\n"
	return fmt.Fprintf(os.Stdout, format, a...)
}

func msg(format string, a ...interface{}) (n int, err error) {
	format = format + "\n"
	return fmt.Fprintf(os.Stdout, format, a...)
}

// A more emphasized version of msg. For section headings.
// FIXME: For now this is just a placeholder. Really do something here
func section(format string, a ...interface{}) (n int, err error) {
	format = format + "\n"
	return fmt.Fprintf(os.Stdout, format, a...)
}
