package gub

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"sort"
	"strings"

	"github.com/rocky/ssa-interp/terminal"
	"code.google.com/p/go-columnize"
)

func Errmsg(format string, a ...interface{}) (n int, err error) {
	if *Highlight {
		format = ansiterm.Colorize("bold", format) + "\n"
	} else {
		format = "** " + format + "\n"
	}
	return fmt.Fprintf(os.Stdout, format, a...)
}

func MsgNoCr(format string, a ...interface{}) (n int, err error) {
	format = format
	return fmt.Fprintf(os.Stdout, format, a...)
}

func Msg(format string, a ...interface{}) (n int, err error) {
	format = format + "\n"
	return fmt.Fprintf(os.Stdout, format, a...)
}

func MsgRaw(msg string) (n int, err error) {
	return fmt.Println(msg)
}

// A more emphasized version of msg. For section headings.
func Section(format string, a ...interface{}) (n int, err error) {
	if *Highlight {
		format = ansiterm.Colorize("bold", format) + "\n"
	} else {
		format = format + "\n" + strings.Repeat("-", len(format)) + "\n"
	}
	return fmt.Fprintf(os.Stdout, format, a...)
}

func PrintSorted(title string, names []string) {
	Section(title + ":")
	sort.Strings(names)
	opts := columnize.DefaultOptions()
	opts.LinePrefix  = "  "
	opts.DisplayWidth = Maxwidth
	columnizedNames := strings.TrimRight(columnize.Columnize(names, opts),
		"\n")
	Msg(columnizedNames)

}

func GetSyntax(syntax ast.Node, fset *token.FileSet) (string, error) {
	buf := new(bytes.Buffer)
	format.Node(buf, fset, syntax)
	if *Highlight {
		highlighted, err := ansiterm.AsTerm([]byte(buf.String()), true)
		if err != nil {
			return "", err
		} else  {
			return string(highlighted), nil
		}
	} else {
		return buf.String(), nil
	}
}

func PrintSyntax(syntax ast.Node, fset *token.FileSet) {
	if str, err := GetSyntax(syntax, fset); err != nil {
		Errmsg(err.Error())
	} else {
		MsgRaw(str)
	}
}

func PrintSyntaxFirstLine(syntax ast.Node, fset *token.FileSet) {
	if str, err := GetSyntax(syntax, fset); err != nil {
		Errmsg(err.Error())
	} else {
		MsgRaw(strings.SplitN(str, "\n", 2)[0])
	}
}
