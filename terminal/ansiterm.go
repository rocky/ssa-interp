//  ansiterm
//
//  Format colored console output.
//
//  Translatted from pygments.go
//  copyright: Copyright 2006-2014 by the Pygments team, see AUTHORS.
//  license: BSD, see LICENSE for details.
package ansiterm

import "fmt"

var esc string
var codes map[string] string
var dark_colors []string
var light_colors []string

func init() {
	dark_colors = []string{"black", "darkred", "darkgreen", "brown", "darkblue",
		"purple", "teal", "lightgray"}
	light_colors = []string{"darkgray", "red", "green", "yellow", "blue",
		"fuchsia", "turquoise", "white"}

	esc = "\x1b["
	codes = make(map[string]string)
	codes[""]          = ""
	codes["reset"]     = esc + "39;49;00m"

	codes["bold"]      = esc + "01m"
	codes["faint"]     = esc + "02m"
	codes["standout"]  = esc + "03m"
	codes["underline"] = esc + "04m"
	codes["blink"]     = esc + "05m"
	codes["overline"]  = esc + "06m"

	for x, d := range(dark_colors) {
		codes[d] = fmt.Sprintf("%s%dm", esc, x+30)
		l := light_colors[x]
		codes[l] = fmt.Sprintf("%s%d;01m", esc, x+30)
	}
	codes["darkteal"]   = codes["turquoise"]
	codes["darkyellow"] = codes["brown"]
	codes["fuscia"]     = codes["fuchsia"]
	codes["white"]      = codes["bold"]
}

func reset_color() string {
    return codes["reset"]
}


func Colorize(color_key string, text string) string {
    return codes[color_key] + text + codes["reset"]
}


//    Format ``text`` with a color and/or some attributes::
//
//        color       normal color
//        *color*     bold color
//        _color_     underlined color
//        +color+     blinking color
//
func ansiformat(attr string, text string) {
    var result []string
    if attr[:1] == attr[len(attr)-1:] && attr[:1] == "+" {
        result = append(result, codes["blink"])
        attr = attr[1:len(attr)-1]
	}
    if attr[:1] == attr[len(attr)-1:] && attr[:1] == "*" {
        result = append(result, codes["bold"])
        attr = attr[1:len(attr)-1]
	}
    if attr[:1] == attr[len(attr)-1:] && attr[:1] == "_" {
        result = append(result, codes["underline"])
        attr = attr[1:len(attr)-1]
	}
    result = append(result, codes[attr])
    result = append(result, text)
    result = append(result, codes["reset"])
}
