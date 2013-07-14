package interp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/rocky/ssa-interp"
)

// Shadows array type
type Array []Value

// Prints in the style of built-in println.
// (More or less; in gc println is actually a compiler intrinsic and
// can distinguish println(1) from println(interface{}(1)).)
// Like toString with these changes:
//   * strings are quoted
//   * separators lists maps are ", " (rather than " ")
//   * nil is "nil" rather than "<nil>"
func toInspect(w io.Writer, v Value) {
	switch v := v.(type) {
	case nil, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, complex64, complex128:
		fmt.Fprintf(w, "%v", v)

	case string:
		fmt.Fprintf(w, "\"%v\"", v)

	case map[Value]Value:
		io.WriteString(w, "map[")
		sep := " "
		for k, e := range v {
			io.WriteString(w, sep)
			sep = ", "
			toInspect(w, k)
			io.WriteString(w, ":")
			toInspect(w, e)
		}
		io.WriteString(w, "]")

	case *hashmap:
		io.WriteString(w, "map[")
		sep := " "
		for _, e := range v.table {
			for e != nil {
				io.WriteString(w, sep)
				sep = ", "
				toInspect(w, e.key)
				io.WriteString(w, ":")
				toInspect(w, e.Value)
				e = e.next
			}
		}
		io.WriteString(w, "]")

	case chan Value:
		fmt.Fprintf(w, "%v", v) // (an address)

	case *Value:
		if v == nil {
			io.WriteString(w, "nil")
		} else {
			fmt.Fprintf(w, "%p", v)
		}

	case iface:
		toInspect(w, v.v)

	case structure:
		io.WriteString(w, "{")
		for i, e := range v {
			if i > 0 {
				io.WriteString(w, ", ")
			}
			toInspect(w, e)
		}
		io.WriteString(w, "}")

	case array:
		io.WriteString(w, "[")
		for i, e := range v {
			if i > 0 {
				io.WriteString(w, ", ")
			}
			toInspect(w, e)
		}
		io.WriteString(w, "]")

	case []Value:
		io.WriteString(w, "[")
		for i, e := range v {
			if i > 0 {
				io.WriteString(w, ", ")
			}
			toInspect(w, e)
		}
		io.WriteString(w, "]")

	case *ssa2.Function, *ssa2.Builtin, *closure:
		fmt.Fprintf(w, "%p", v) // (an address)

	case rtype:
		io.WriteString(w, v.t.String())

	case tuple:
		// Unreachable in well-formed Go programs
		io.WriteString(w, "(")
		for i, e := range v {
			if i > 0 {
				io.WriteString(w, ", ")
			}
			toInspect(w, e)
		}
		io.WriteString(w, ")")

	default:
		fmt.Fprintf(w, "<%T>", v)
	}
}

// Similar to ToString but using toInspect
func ToInspect(v Value) string {
	var b bytes.Buffer
	toInspect(&b, v)
	return b.String()
}
