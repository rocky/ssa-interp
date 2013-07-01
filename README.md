gub and tortoise - A Go SSA Debugger and Interpreter
============================================================================

This projects provides a debugger for the SSA-builder and interpeter from http://code.google.com/go.tools/ssa .

For now, some of the SSA optimizations however have been removed since our focus is on debugging and extending for interactive evaluation. Those optimization can be added back in when we have a good handle on debugging and interactive evaluation.

Setup
-----

	git clone http://github.com/rocky/go-ssa-interp
	cd go-ssa-interp
	make
	# copy tortoise and gub.sh somewhere in your PATH
	gub.sh -- *go-program* [*program-opts*...]
