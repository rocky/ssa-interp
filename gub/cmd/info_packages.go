// Copyright 2013-2014 Rocky Bernstein.

// info packages
//
// Prints package information

package gubcmd

import (
	"reflect"
	"sort"
	"strings"
	"code.google.com/p/go-columnize"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoPackageSubcmd,
		Help: `info program

Prints package name and path information about :
`,
		Min_args: 0,
		Max_args: -1,  // Max_args < 0 means an arbitrary number
		Short_help: "Information about packages",
		Name: "package",
	})
}

func printReflectMap(title string, m map[string] reflect.Value) {
	if len(m) > 0 {
		list := []string {}
		for item := range m {
			list = append(list, item)
		}
		gub.PrintSorted(title, list)
	}
}

func printReflectTypeMap(title string, m map[string] reflect.Type) {
	if len(m) > 0 {
		list := []string {}
		for item := range m {
			list = append(list, item)
		}
		gub.PrintSorted(title, list)
	}
}

// InfoPackageCommand implements the command:
//    info package [*name* ]
// which show information about a package or lists all packages.
func InfoPackageSubcmd(args []string) {
	if len(args) > 2 {
		for _, pkg_name := range args[2:len(args)] {
			if pkg := gub.Program().PackagesByName[pkg_name]; pkg != nil {
				gub.Msg("Package %s: \"%s\"", pkg_name, pkg.Object.Path())
				gub.Section("Package members:")
				var names []string
				for k, _ := range pkg.Members {
					names = append(names, k)
				}
				sort.Strings(names)
				opts := columnize.DefaultOptions()
				opts.DisplayWidth = gub.Maxwidth
				opts.LinePrefix  = "  "
				mems := strings.TrimRight(columnize.Columnize(names, opts),
					"\n")
				gub.Msg(mems)

			} else {
				gub.Errmsg("Package %s not imported", pkg_name)
			}
		}
	} else {
		pkgNames := []string {}
		for pkg := range gub.Program().PackagesByName {
			pkgNames = append(pkgNames, pkg)
		}
		gub.PrintSorted("All imported packages", pkgNames)
	}
}
