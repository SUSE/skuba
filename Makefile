GO ?= go
GO_MD2MAN ?= go-md2man


VERSION    := $(shell cat VERSION)
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
CAASPCTL_LDFLAGS = -ldflags "-X=suse.com/caaspctl/internal/app/caaspctl.Version=$(VERSION) \
                             -X=suse.com/caaspctl/internal/app/caaspctl.Commit=$(COMMIT) \
                             -X=suse.com/caaspctl/internal/app/caaspctl.BuildDate=$(BUILD_DATE)"
.PHONY: all
all: build

.PHONY: build
build:
	$(GO) install $(CAASPCTL_LDFLAGS) suse.com/caaspctl/cmd/...

MANPAGES_MD := $(wildcard docs/man/*.md)
MANPAGES    := $(MANPAGES_MD:%.md=%)

docs/man/%.1: docs/man/%.1.md
	$(GO_MD2MAN) -in $< -out $@

.PHONY: docs
docs: $(MANPAGES)

.PHONY: staging
staging:
	$(GO) install $(CAASPCTL_LDFLAGS) -tags staging suse.com/caaspctl/cmd/...

.PHONY: release
release:
	$(GO) install $(CAASPCTL_LDFLAGS) -tags release suse.com/caaspctl/cmd/...
