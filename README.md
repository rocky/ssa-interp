gub and tortoise - A Go SSA Debugger and Interpreter
============================================================================

This projects provides a debugger for the SSA-builder and interpeter from http://code.google.com/p/go/source/checkout?repo=tools .

For now, some of the SSA optimizations however have been removed since our focus is on debugging and extending for interactive evaluation. Those optimization can be added back in when we have a good handle on debugging and interactive evaluation.

Setup
-----

* Make sure our GO environment is setup, e.g. *$GOBIN*, *$GOPATH*, ...
* Make sure you have go 1.1.1 installed. For go 1.2 use the go-1.2 branch.

```
   go get github.com/rocky/ssa-interp
   cd <place where ssa-interp/src copied>
   make
   cp tortroise gub.sh  $GOBIN/
```

Running the debugger:

```
  gub.sh -- *go-program* [-- *program-opts*...]
```

Or the interpreter, *tortoise*:

```
  tortoise -run *go-program* [-- *program-opts*..]
```

See Also
--------

* [What's left to do?](https://github.com/rocky/ssa-interp/wiki/What%27s-left-to-do%3F)
* [Cool things](https://github.com/rocky/ssa-interp/wiki/Cool-things)
