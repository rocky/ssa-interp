# Whatever it is you want to do, it should be forwarded to the
# to top-level irectories
.PHONY: all check test check-quick

all:
	$(MAKE) -C .. tortoise

#: Run quick tests
check-quick: ../tortoise
	go test -test.short

#: Run all tests
check: ../tortoise
	go test

#: Same as check
test:	check

%:
	$(MAKE) -C .. $@
