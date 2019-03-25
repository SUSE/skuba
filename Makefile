GO ?= go
GO_MD2MAN ?= go-md2man

.PHONY: all
all: build

.PHONY: build
build:
	$(GO) install suse.com/caaspctl/cmd/...

MANPAGES_MD := $(wildcard docs/man/*.md)
MANPAGES    := $(MANPAGES_MD:%.md=%)

docs/man/%.1: docs/man/%.1.md
	$(GO_MD2MAN) -in $< -out $@

.PHONY: docs
docs: $(MANPAGES)
