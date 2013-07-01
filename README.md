gub and tortoise - A Go SSA Debugger and Interpreter
============================================================================

This projects provides a debugger for the SSA-builder and interpeter from http://code.google.com/p/go/source/checkout?repo=tools .

For now, some of the SSA optimizations however have been removed since our focus is on debugging and extending for interactive evaluation. Those optimization can be added back in when we have a good handle on debugging and interactive evaluation.

Setup
-----

Make sure you have go 1.1.1 or later installed.

     go get github.com/rocky/ssa-interp

     cd <place where ssa-interp/src copied>
     cp tortroise gub.sh  $GOBIN/

Running the debugger:

     gub.sh -- *go-program* [-- *program-opts*...]

Or the interpreter, tortoise:

     tortoise -run *go-program* [-- *program-opts*..]
