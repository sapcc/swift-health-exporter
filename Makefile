PKG     = github.com/sapcc/swift-health-exporter
PREFIX := /usr

GO            := GOBIN=$(CURDIR)/build go
GO_BUILDFLAGS :=
GO_LDFLAGS    := -s -w

all: build/swift-health-exporter

build/swift-health-exporter: FORCE | build
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)'

install: FORCE all
	install -D -m 0755 build/swift-health-exporter "$(DESTDIR)$(PREFIX)/bin/swift-health-exporter"

# which packages to test with static checkers?
GO_ALLPKGS := $(shell $(GO) list $(GO_BUILDFLAGS) $(PKG)/...)
# which packages to test with `go test`?
GO_TESTPKGS := $(shell $(GO) list $(GO_BUILDFLAGS) -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...)

check: FORCE all static-check build/cover.html
	@printf "\e[1;32m>> All tests successful.\e[0m\n"

static-check: FORCE
	@if ! hash golint 2>/dev/null; then printf "\e[1;36m>> Installing golint...\e[0m\n"; GO111MODULE=off go get -u golang.org/x/lint/golint; fi
	@printf "\e[1;36m>> gofmt\e[0m\n"
	@if s="$$(gofmt -s -l *.go cmd pkg 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
	@printf "\e[1;36m>> golint\e[0m\n"
	@if s="$$(golint . && find cmd pkg -type d -exec golint {} \; 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi
	@printf "\e[1;36m>> go vet\e[0m\n"
	@$(GO) vet $(GO_BUILDFLAGS) $(GO_ALLPKGS)

# detailed unit test run (incl. test coverage)
build/cover.out: FORCE build/mock-tools | build
	@printf "\e[1;36m>> go test with coverage\e[0m\n"
	$(GO) test $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' -failfast -race -covermode=atomic -coverpkg=$(PKG)/collector -coverprofile=$@ $(GO_TESTPKGS)
build/cover.html: build/cover.out
	$(GO) tool cover -html $< -o $@

# quick unit test run
quick-check: FORCE build/mock-tools | build
	@printf "\e[1;36m>> go test\e[0m\n"
	$(GO) test $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' $(GO_TESTPKGS)
	@printf "\e[1;32m>> Unit tests successful.\e[0m\n"

build/mock-tools: FORCE | build
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)/test/cmd/mock-swift-dispersion-report'
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)/test/cmd/mock-swift-dispersion-report-with-errors'
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)/test/cmd/mock-swift-recon'
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)/test/cmd/mock-swift-recon-with-errors'

build:
	mkdir $@

clean: FORCE
	rm -rf -- build/

vendor: FORCE
	$(GO) mod tidy -v
	$(GO) mod vendor
	$(GO) mod verify

.PHONY: FORCE
