package gub

import (
	"fmt"
	"os"
)

func Errmsg(format string, a ...interface{}) (n int, err error) {
	format = "** " + format + "\n"
	return fmt.Fprintf(os.Stdout, format, a...)
}

func Msg(format string, a ...interface{}) (n int, err error) {
	format = format + "\n"
	return fmt.Fprintf(os.Stdout, format, a...)
}

// A more emphasized version of msg. For section headings.
// FIXME: For now this is just a placeholder. Really do something here
func Section(format string, a ...interface{}) (n int, err error) {
	format = format + "\n"
	return fmt.Fprintf(os.Stdout, format, a...)
}
