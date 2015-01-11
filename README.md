gub and tortoise - A Go SSA Debugger and Interpreter
============================================================================

[![Build Status](https://travis-ci.org/rocky/ssa-interp.png)](https://travis-ci.org/rocky/ssa-interp)

This project modifies the [Go SSA interpreter](https://godoc.org/golang.org/x/tools/go/ssa/interp) to support a debugger. We provide the debugger here as well.

Setup
-----

* Make sure our Go environment is setup, e.g. *$GOBIN*, *$GOPATH*, ...
* Make sure you have a 1.4 Go version installed. For Go 1.2.2 use the go1.2.2 branch, for Go 1.1.1, use the go1.1 branch.


```
    bash < <(curl -s https://raw.github.com/rocky/ss-ainterp/master/gub-installer)
```

If this doesn't work here is what the above does:

```
   go get github.com/rocky/ssa-interp
   cd $GOBIN/src/github.com/rocky/ssa-interp
   make install
```

Running
-------

Running the debugger:

```
  gub.sh -- *go-program* [-- *program-opts*...]
  tortoise -run -interp=S *go-program* [-- *program-opts*..]  # Same as above
```

Or just the interpreter, *tortoise*:

```
  tortoise -run *go-program* [-- *program-opts*..]
```

and now with a test suite, *columnize*:

```
  tortoise -run -interp=S -test columnize
```

See Also
--------

* [What's left to do?](https://github.com/rocky/ssa-interp/wiki/What%27s-left-to-do%3F)
* [Cool things](https://github.com/rocky/ssa-interp/wiki/Cool-things)
* [go-play](http://code.google.com/p/go-play): A locally-run HTML5 web interface for experimenting with Go code
* [go-fish](https://github.com/rocky/go-fish): Yet another Go REPL

[![endorse rocky](https://api.coderwall.com/rocky/endorsecount.png)](https://coderwall.com/rocky)
