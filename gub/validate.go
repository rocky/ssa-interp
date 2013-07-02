// Copyright 2013 Rocky Bernstein.
// command argument-validation routines
package gub

import "strconv"

func argCountOK(min int, max int, args [] string) bool {
	l := len(args)-1 // strip command name from count
	if (l < min) {
		errmsg("Too few args; need at least %d, got %d", min, l)
		return false
	} else if (l > max) {
		errmsg("Too many args; need at most %d, got %d", max, l)
		return false
	}
	return true
}

type NumError struct {
	bogus bool
}

func (e *NumError) Error() string {
	return "generic error"
}
var genericError = &NumError{bogus: true}

func getInt(arg string, what string, min int, max int) (int, error) {
	errmsg_fmt := "Expecting integer " + what + "; got '%s'."
	i, err := strconv.Atoi(arg)
	if err != nil {
		errmsg(errmsg_fmt, arg)
		return 0, err
	}
	if i < min {
		errmsg("Expecting integer value to be at least %d; got %d.",
			min, i)
        return 0, genericError
	} else if i > max {
        errmsg("Expecting integer value to be at most %d; got %d.",
			max, i)
        return 0, genericError
	}
	return i, nil
}
