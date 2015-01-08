# Comments starting with #: below are remake GNU Makefile comments. See
# https://github.com/rocky/remake/wiki/Rake-tasks-for-gnu-make

.PHONY: all builder interp check test check-quick check-interp check-interp-quick test-quick clean

#: Same as tortoise
all: tortoise

#: The front-end to the builder, interpreter, and debugger
tortoise: interp builder gub trepan
	(cd cmd && go build tortoise.go)

#: Build the SSA Builder
builder:
	go install

#: Build the interpeter
interp: builder big
	(cd interp && go install)

#: Build our replacement for math/big
big:
	(cd interp/big && go install)

#: Build the debugger
gub: interp
	(cd gub && go install)
	(cd gub/cmd && go install)

#: Same as "check"
test: check

#: Run all tests (quick and interpreter)
check:
	go test -i && go test
	(cd interp && go test -i && go test)
	(cd gub && go test -i && go test)

#: Run quick tests
check-quick:
	go test -i && go test
	(cd interp && go test -i && go test -test.short)
	(cd gub && go test -i && go test -test.short)

#: Same as check-quick
test-quick: check-quick

#: Longer interpreter tests
check-interp:
	(cd interp && go test -i && go test)


#: Shorter interpreter tests
check-interp-quick:
	(cd interp && go test -i && go test -test.short)

#: all debugger tests
check-gub:
	(cd gub && go test -i && go test)

#: Shorter debugger tests
check-gub-quick:
	(cd gub && go test -i && go test -test.short)

#: Remove derived files.
clean:
	for file in tortoise ; do \
		[ -e $$file ] && rm $$file; \
	done
