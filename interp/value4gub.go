package interp

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

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
		fmt.Fprintf(w, strconv.QuoteToASCII(v))


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
// Note: we can't use a method because the receiver is an interface type.
func ToInspect(v Value) string {
	var b bytes.Buffer
	toInspect(&b, v)
	return b.String()
}

// Returns a string representation of the types of interp.Value
// Note: we can't use a method becasue the receiver is an interface type.
func Type(v Value) string {
	switch v.(type) {
	case nil:
		return "nil"
	case bool:
		return "bool"
	case int:
		return "int"
	case int8:
		return "int8"
	case int16:
		return "int16"
	case int32:
		return "int32"
	case int64:
		return "int64"
	case uint:
		return "uint"
	case uint8:
		return "uint8"
	case uint16:
		return "uint16"
	case uint32:
		return "uint32"
	case uint64:
		return "uint64"
	case uintptr:
		return "uintptr"
	case float32:
		return "float32"
	case float64:
		return "float64"
	case complex64:
		return "complex64"
	case complex128:
		return "complex128"
	case string:
		return "string"
	case map[Value]Value:
		return "map[Value]Value"
	case *hashmap:
		return "*hashmap"
	case chan Value:
		return "chan Value"
	case *Value:
		return "*Value"
	case iface:
		return "iface"
	case structure:
		return "structure"
	case array:
		return "array"
	case []Value:
		return "[]Value"
	case *ssa2.Function:
		return "*ssa2.Function"
	case *ssa2.Builtin:
		return "*ssa2.Builtin"
	case *closure:
		return "*closure"
	case rtype:
		return "rtype"
	case tuple:
		return "tuple"
	default:
		return "?"
	}
}
