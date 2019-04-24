GO ?= go
GO_MD2MAN ?= go-md2man
LN = ln
RM = rm

GOBINPATH  := $(GOPATH)/bin
VERSION    := $(shell cat VERSION)
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
CAASPCTL_LDFLAGS = -ldflags "-X=github.com/SUSE/caaspctl/internal/app/caaspctl.Version=$(VERSION) \
                             -X=github.com/SUSE/caaspctl/internal/app/caaspctl.Commit=$(COMMIT) \
                             -X=github.com/SUSE/caaspctl/internal/app/caaspctl.BuildDate=$(BUILD_DATE)"
.PHONY: all
all: build

.PHONY: build
build:
	$(GO) build $(CAASPCTL_LDFLAGS) -o $(GOBINPATH)/caaspctl ./cmd/caaspctl
	$(RM) -f $(GOBINPATH)/kubectl-caasp
	$(LN) -s $(GOBINPATH)/caaspctl $(GOBINPATH)/kubectl-caasp

MANPAGES_MD := $(wildcard docs/man/*.md)
MANPAGES    := $(MANPAGES_MD:%.md=%)

docs/man/%.1: docs/man/%.1.md
	$(GO_MD2MAN) -in $< -out $@

.PHONY: docs
docs: $(MANPAGES)

.PHONY: staging
staging:
	$(GO) build $(CAASPCTL_LDFLAGS) -o $(GOBINPATH)/caaspctl -tags staging  ./cmd/caaspctl
	$(RM) -f $(GOBINPATH)/kubectl-caasp
	$(LN) -s $(GOBINPATH)/caaspctl $(GOBINPATH)/kubectl-caasp

.PHONY: release
release:
	$(GO) build $(CAASPCTL_LDFLAGS) -o $(GOBINPATH)/caaspctl -tags release ./cmd/caaspctl
	$(RM) -f $(GOBINPATH)/kubectl-caasp
	$(LN) -s $(GOBINPATH)/caaspctl $(GOBINPATH)/kubectl-caasp
