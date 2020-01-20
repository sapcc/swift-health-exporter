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

# which packages to test with static checkers?
GO_ALLPKGS := $(shell go list $(GO_BUILDFLAGS) $(PKG)/...)
# which packages to test with `go test`?
GO_TESTPKGS := $(shell go list $(GO_BUILDFLAGS) -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' $(PKG)/...)
# which packages to measure coverage for?
GO_COVERPKGS := $(shell go list $(GO_BUILDFLAGS) $(PKG) $(PKG)/collector)
# output files from `go test`
GO_COVERFILES := $(patsubst %,build/%.cover.out,$(subst /,_,$(GO_TESTPKGS)))

# down below, I need to substitute spaces with commas; because of the syntax,
# I have to get these separators from variables
space := $(null) $(null)
comma := ,

check: FORCE all static-check build/cover.html
	@printf "\e[1;32m>> All tests successful.\e[0m\n"

static-check: FORCE
	@if ! hash golint 2>/dev/null; then printf "\e[1;36m>> Installing golint...\e[0m\n"; go get -u golang.org/x/lint/golint; fi
	@printf "\e[1;36m>> gofmt\e[0m\n"
	@if s="$$(gofmt -s -l *.go cmd pkg 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
	@printf "\e[1;36m>> golint\e[0m\n"
	@if s="$$(golint . && find cmd pkg -type d -exec golint {} \; 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi
	@printf "\e[1;36m>> go vet\e[0m\n"
	@$(GO) vet $(GO_BUILDFLAGS) $(GO_ALLPKGS)

# detailed unit test run (incl. test coverage)
build/%.cover.out: FORCE build/mock-tools
	@printf "\e[1;36m>> go test $(subst _,/,$*)\e[0m\n"
	$(GO) test $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' -coverprofile=$@ -covermode=count -coverpkg=$(subst $(space),$(comma),$(GO_COVERPKGS)) $(subst _,/,$*)
build/cover.out: $(GO_COVERFILES)
	$(GO) run $(GO_BUILDFLAGS) test/cmd/gocovcat/main.go $(GO_COVERFILES) > $@
build/cover.html: build/cover.out
	$(GO) tool cover -html $< -o $@

# quick unit test run
quick-check: FORCE all build/mock-tools $(addprefix quick-check-,$(subst /,_,$(GO_TESTPKGS)))
quick-check-%:
	@printf "\e[1;36m>> go test $(subst _,/,$*)\e[0m\n"
	$(GO) test $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' $(subst _,/,$*)
	@printf "\e[1;32m>> Unit tests successful.\e[0m\n"

build/mock-tools: FORCE
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)/test/cmd/mock-swift-dispersion-report'
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)/test/cmd/mock-swift-dispersion-report-with-errors'
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)/test/cmd/mock-swift-recon'
	$(GO) install $(GO_BUILDFLAGS) -ldflags '$(GO_LDFLAGS)' '$(PKG)/test/cmd/mock-swift-recon-with-errors'

clean: FORCE
	rm -rf -- build

vendor: FORCE
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: FORCE
