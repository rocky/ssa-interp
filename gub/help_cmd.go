// Copyright 2013 Rocky Bernstein.

package gub

import (
	"sort"
	"strings"
	"code.google.com/p/go-columnize"
)

func init() {
	name := "help"
	cmds[name] = &CmdInfo{
		fn: HelpCommand,
		help: `List of commands:
Execution running --
  s: step in
  n: next or step over
  fin: finish or step out
  c: continue

Inspecting --
  disasm [*fn*]  : disassemble function
  env            : dump frame environment
  locs           : show breakpoint locations
  local [*name*] : show local variable info
  global [*name*]: show global variable info
  param [*name*] : show function parameter info
  whatis *name*  : show information about name
  locs           : show all stopping locations

Breakpoints --

  break : list breakpoints
  break line [column] : break at this line (and column)
                      : run locs for a list
  break function      : break at function

  enable bpnum [bpnum..]    : enable breakpoint
  disable bpnum [bpnum...]  : disable breakpoint
  delete bpnum              : delete breakpoint

Tracing --
  +: add instruction tracing
  -: remove instruction tracing

Stack:
  bt [*max*]  : print a backtrace (at most max entries)
  frame *num* : switch stack frame
  gor [*num*] : show goroutine stack (for num)
  up *num*    : switch to a newer frame
  down *num*  : switch to a older frame

Other:
  ?: this help
  q: quit
`,
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("support", name)
	AddAlias("?", name)
	// Down the line we'll have abbrevs
	AddAlias("h", name)
}

func HelpCommand(args []string) {
	if len(args) == 1 {
		msg(cmds["help"].help)
	} else {
		what := args[1]
		cmd := lookupCmd(what)
		if what == "*" {
			var names []string
			for k, _ := range cmds {
				names = append(names, k)
			}
			section("All command names:")
			sort.Strings(names)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = maxwidth
			mems := strings.TrimRight(columnize.Columnize(names, opts),
				"\n")
			msg(mems)
		} else if info := cmds[cmd]; info != nil {
			msg(info.help)
			if len(info.aliases) > 0 {
				msg("Aliases: %s",
					strings.Join(info.aliases, ", "))
			}
		} else if cmds := categories[what]; len(cmds) > 0 {
			section("Commands in class: %s", what)
			sort.Strings(cmds)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = maxwidth
			mems := strings.TrimRight(columnize.Columnize(cmds, opts),
				"\n")
			msg(mems)
		} else {
			errmsg("Can't find help for %s", what)
		}
	}
}
