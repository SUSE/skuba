GO ?= go
GOFMT ?= gofmt
TERRAFORM ?= $(shell which terraform 2>/dev/null || which true 2>/dev/null)
GO_MD2MAN ?= go-md2man
LN = ln
RM = rm

GOBINPATH    := $(shell $(GO) env GOPATH)/bin
VERSION      := $(shell cat VERSION)
COMMIT       := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE   := $(shell date +%Y%m%d-%H:%M:%S)
TAGS         := development
CAASPCTL_LDFLAGS = -ldflags "-X=github.com/SUSE/caaspctl/internal/app/caaspctl.Version=$(VERSION) \
                             -X=github.com/SUSE/caaspctl/internal/app/caaspctl.Commit=$(COMMIT) \
                             -X=github.com/SUSE/caaspctl/internal/app/caaspctl.BuildDate=$(BUILD_DATE)"

CAASPCTL_DIRS    = cmd pkg internal test

# go source files, ignore vendor directory
CAASPCTL_SRCS = $(shell find $(CAASPCTL_DIRS) -type f -name '*.go')

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

.PHONY: lint
lint:
	$(GO) tool vet ${CAASPCTL_SRCS}
	test -z `$(GOFMT) -l $(CAASPCTL_SRCS)` || { $(GOFMT) -d $(CAASPCTL_SRCS) && false; }
	$(TERRAFORM) fmt -check=true -write=false -diff=true ci/infra
	find ci -type f -name "*.sh" | xargs bashate

.PHONY: suse-package
suse-package:
	ci/packaging/suse/rpmfiles_maker.sh $(VERSION)

.PHONY: suse-changelog
suse-changelog:
	ci/packaging/suse/changelog_maker.sh "$(CHANGES)"
