GO ?= go
GO_MD2MAN ?= go-md2man


VERSION    := $(shell cat VERSION)
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
CAASPCTL_LDFLAGS = -ldflags "-X=github.com/SUSE/caaspctl/internal/app/caaspctl.Version=$(VERSION) \
                             -X=github.com/SUSE/caaspctl/internal/app/caaspctl.Commit=$(COMMIT) \
                             -X=github.com/SUSE/caaspctl/internal/app/caaspctl.BuildDate=$(BUILD_DATE)"

CAASPCTL_DIRS    = cmd pkg internal test

# go source files, ignore vendor directory
CAASPCTL_SRCS = $(shell find $(CAASPCTL_DIRS) -type f -name '*.go')

.PHONY: all
all: build

.PHONY: build
build:
	$(GO) install $(CAASPCTL_LDFLAGS) ./cmd/...

MANPAGES_MD := $(wildcard docs/man/*.md)
MANPAGES    := $(MANPAGES_MD:%.md=%)

docs/man/%.1: docs/man/%.1.md
	$(GO_MD2MAN) -in $< -out $@

.PHONY: docs
docs: $(MANPAGES)

.PHONY: staging
staging:
	$(GO) install $(CAASPCTL_LDFLAGS) -tags staging ./cmd/...

.PHONY: release
release:
	$(GO) install $(CAASPCTL_LDFLAGS) -tags release ./cmd/...

.PHONY: vet
vet:
	$(GO) tool vet ${CAASPCTL_SRCS} 
