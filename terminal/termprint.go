// Printing to an ANSI-type terminal for github.com/sourcegraph/syntaxhighlight
package ansiterm

import (
	"bytes"
	"io"
	// "github.com/shurcooL/go/highlight_go"
	"github.com/sourcegraph/syntaxhighlight"
)

type TermPrinter TermConfig

const (
	Bold       = iota+10
	Underline
	Italic
	Strike
	BoldItalic
	Builtin
	Function
	Variable
	Operator
	Number
)

type TermConfig struct {
	Comment       string
	Keyword       string
	Type          string
	Builtin       string
	Function      string
	Variable      string
	Operator      string
	String        string
	Decimal       string
}

// DefaultTermConfig's class names match those of pygments
// (https://pygments.org/).
var LightTermConfig = TermConfig{
	Comment:       "lightgray", //
	Keyword:       "darkblue",  // blue
	Type:          "teal",      // turquoise
	Builtin:       "teal",      // turquoise
	Function:      "darkgreen", // green
	Variable:      "darkred",   // red
	Operator:      "purple",    // fuchsia
	String:        "brown",     // brown
	Decimal:       "darkblue",  // blue
}
var DarkTermConfig = TermConfig{
	Comment:       "darkgray",
	Keyword:       "blue",
	Type:          "turquoise",
	Builtin:       "turquoise",
	Function:      "green",
	Variable:      "red",
	Operator:      "fuscia",
	String:        "brown",
	Decimal:        "blue",
}

func (c TermConfig) class(kind syntaxhighlight.Kind) string {
	// println(kind)
	switch kind {
	case syntaxhighlight.Keyword:
		return c.Keyword
	case syntaxhighlight.Comment:
		return c.Comment
	case syntaxhighlight.Type:
		return c.Type
	// case syntaxhighlight.Builtin:
	// 	return c.Builtin
	// case syntaxhighlight.Function:
	// 	return c.Function
	// case syntaxhighlight.Variable:
	// 	return c.Variable
	// case syntaxhighlight.Operator:
	// 	return c.Operator
	case syntaxhighlight.String:
		return c.String
	case syntaxhighlight.Decimal:
		return c.Decimal
	case syntaxhighlight.Punctuation:
		return ""
	case syntaxhighlight.Plaintext:
		return ""
	}
	return ""
}

func (p TermPrinter) Print(w io.Writer, kind syntaxhighlight.Kind, tokText string) error {
	class := ((TermConfig)(p)).class(kind)
	if class != "" {
		// println(class)
		_, err := io.WriteString(w, Colorize(class, tokText))
		if err != nil {
			return err
		}
	} else {
		io.WriteString(w, tokText)
	}
	return nil
}


func AsTerm(src []byte, isDark bool) ([]byte, error) {
	var buf bytes.Buffer
	config := LightTermConfig
	if isDark { config = DarkTermConfig }
	err := syntaxhighlight.Print(syntaxhighlight.NewScanner(src), &buf,
		TermPrinter(config))
	// err := highlight_go.Print(src, &buf, TermPrinter(config))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
