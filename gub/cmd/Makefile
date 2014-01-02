# Whatever it is you want to do, it should be forwarded to the
# to top-level directories
.PHONY: check check-quick test test-quick all

all:
	$(MAKE) -C ../.. tortoise

#: Run quick tests
check-quick: ../../tortoise
	go test -test.short

#: Run all tests
check: ../../tortoise
	cd .. && go test

#: Same as check
test:	check

#: Same as check-quick
test-quick: check-quick

%:
	$(MAKE) -C ../.. $@
