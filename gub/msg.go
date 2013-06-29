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
