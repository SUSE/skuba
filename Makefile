GO        ?= go
BINDIR    := $(CURDIR)/bin

.PHONY: all
all: build

.PHONY: build
build:
	GOBIN=$(BINDIR) $(GO) install suse.com/caaspctl/cmd/...
