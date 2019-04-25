GO ?= go
GO_MD2MAN ?= go-md2man
LN = ln
RM = rm

GOBINPATH    := $(shell $(GO) env GOPATH)/bin
VERSION      := $(shell cat VERSION)
COMMIT       := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE   := $(shell date +%Y%m%d-%H:%M:%S)
TAGS         := devel
CAASPCTL_LDFLAGS = -ldflags "-X=github.com/SUSE/caaspctl/internal/app/caaspctl.Version=$(VERSION) \
                             -X=github.com/SUSE/caaspctl/internal/app/caaspctl.Commit=$(COMMIT) \
                             -X=github.com/SUSE/caaspctl/internal/app/caaspctl.BuildDate=$(BUILD_DATE)"

.PHONY: all
all: install

.PHONY: build
build:
	$(GO) install $(CAASPCTL_LDFLAGS) -tags $(TAGS) ./cmd/...

MANPAGES_MD := $(wildcard docs/man/*.md)
MANPAGES    := $(MANPAGES_MD:%.md=%)

docs/man/%.1: docs/man/%.1.md
	$(GO_MD2MAN) -in $< -out $@

.PHONY: docs
docs: $(MANPAGES)

.PHONY: install
install: build
	  $(RM) -f $(GOBINPATH)/kubectl-caasp
	  $(LN) -s $(GOBINPATH)/caaspctl $(GOBINPATH)/kubectl-caasp

.PHONY: staging
staging:
	 make TAGS=staging install

.PHONY: release
release:
	make TAGS=release install
