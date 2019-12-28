PKG     = github.com/sapcc/swift-health-exporter
PREFIX := /usr

# NOTE: This repo uses Go modules, and uses a synthetic GOPATH at
# $(CURDIR)/.gopath that is only used for the build cache. $GOPATH/src/ is
# empty.
GO            := GOPATH=$(CURDIR)/.gopath GOBIN=$(CURDIR)/build go
GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w

all: build/swift-health-exporter

build/swift-health-exporter: FORCE
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)'

install: FORCE all
	install -D -m 0755 build/swift-health-exporter "$(DESTDIR)$(PREFIX)/bin/swift-health-exporter"

clean: FORCE
	rm -rf -- build

vendor: FORCE
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: FORCE
