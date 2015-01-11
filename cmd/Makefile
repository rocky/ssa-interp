# Whatever it is you want to do other run demos, it should be forwarded to the
# to top-level directory
.PHONY: all tortoise install

#: the default target - same as running "check"
all:
	$(MAKE) -C .. tortoise

tortoise: tortoise.go
	go build tortoise.go

#: install this
install: tortoise gub.sh
	cp tortoise gub.sh $${GOBIN}

%:
	$(MAKE) -C .. $@
