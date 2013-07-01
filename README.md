gub and tortoise - A Go SSA Debugger and Interpreter
============================================================================

This projects provides a debugger for the SSA-builder and interpeter from http://code.google.com/p/go/source/checkout?repo=tools .

For now, some of the SSA optimizations however have been removed since our focus is on debugging and extending for interactive evaluation. Those optimization can be added back in when we have a good handle on debugging and interactive evaluation.

Setup
-----

Make sure you have go 1.1.1 installed.


    cd $GOPATH/src  # or someplace in your $
    git clone http://github.com/rocky/ssa-interp
    cd ssa-interp
    make
    # copy tortoise and gub.sh somewhere in your PATH
    gub.sh -- *go-program* [*program-opts*...]
