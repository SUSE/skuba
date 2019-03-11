GO ?= go

.PHONY: all
all: build

.PHONY: build
build:
	$(GO) install suse.com/caaspctl/cmd/...
