# Convenience wrappers around the commands in AGENTS.md. Run `make help`.

SHELL  := bash
GO     ?= go
PKGS   ?= ./...
DEMO   ?= button
OUTDIR ?= tmp/bin

.DEFAULT_GOAL := help

.PHONY: help all build test test-race vet fmt fmt-check fix fix-check lint \
	check ci build-windows vet-windows demos run-demo bench fuzz apidiff release \
	clean distclean

help: ## List the goals
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

all: build test ## Build and run the short tests

build: ## Build everything for the host platform
	$(GO) build $(PKGS)

test: ## Short unit tests (GUI tests start a private Xvfb when installed)
	$(GO) test -short $(PKGS)

test-race: ## Short tests with the race detector, as CI does
	$(GO) test -short -race $(PKGS)

vet: ## go vet
	$(GO) vet $(PKGS)

# Go files git does not ignore, and the buildable package directories among
# them, so the gitignored scratch space (tmp/, tk/, tcl/) is left out: its
# programs may not build and would show up as errors. Directories, not import
# paths, as golangci-lint wants; go list drops packages for other platforms.
GOFILES ?= $(shell git ls-files -co --exclude-standard '*.go')
GODIRS  ?= $(shell $(GO) list -f '{{.Dir}}' ./... 2>/dev/null | sed "s|^$(CURDIR)|.|" | \
	grep -Fxf <(printf '%s\n' $(GOFILES) | xargs -n1 dirname | sed 's|^[^.]|./&|' | sort -u))

fmt: ## Format all tracked Go files
	gofmt -w $(GOFILES)

fmt-check: ## Fail if any tracked file needs gofmt
	@out=$$(gofmt -l $(GOFILES)); if [ -n "$$out" ]; then echo "needs gofmt:"; echo "$$out"; exit 1; fi

fix: ## Apply the go fix modernizers
	$(GO) fix $(GODIRS)

fix-check: ## Fail if go fix would change anything (CI gate)
	$(GO) fix -diff $(GODIRS)

check-docs: ## Fail if AGENTS.md, the skills or the scripts README name a path that is gone
	bash scripts/check_docs.sh

lint: ## golangci-lint (CI only gates changed lines)
	golangci-lint run --allow-parallel-runners $(GODIRS)

check: vet fmt-check fix-check vet-windows check-docs ## The static checks CI gates on

ci: check test-race ## Roughly what CI runs on Linux

build-windows: ## Cross-compile every demo to tmp/bin/windows (pure Go)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -o $(OUTDIR)/windows/ ./demos/...

vet-windows: ## Vet the Windows backend from Linux
	CGO_ENABLED=0 GOOS=windows $(GO) vet $(PKGS)

demos: ## Build every demo for the host to tmp/bin/demos
	$(GO) build -o $(OUTDIR)/demos/ ./demos/...

run-demo: ## Run one demo: make run-demo DEMO=button
	$(GO) run ./demos/$(DEMO)

bench: ## Run benchmarks without tests
	$(GO) test $(PKGS) -run '^$$' -bench .

fuzz: ## Fuzz the binding parser: make fuzz FUZZTIME=30s
	$(GO) test ./bind/ -run '^$$' -fuzz '^FuzzParse$$' -fuzztime $(or $(FUZZTIME),30s)

apidiff: ## Exported API changes since the last tag: make apidiff [BASE=v0.2.0]
	bash scripts/apidiff.sh $(BASE)

release: ## Tag a release from CHANGELOG.md: make release VERSION=v0.2.0 [PUSH=1]
	@[ -n "$(VERSION)" ] || { echo "usage: make release VERSION=vX.Y.Z [PUSH=1]"; exit 1; }
	bash scripts/release.sh $(VERSION) $(if $(PUSH),--push)

clean: ## Remove built binaries
	rm -rf $(OUTDIR)/demos $(OUTDIR)/windows

distclean: clean ## clean, plus Go build/test/fuzz caches and generated tmp/ files
	$(GO) clean -cache -testcache -fuzzcache
	rm -rf $(OUTDIR) tmp/fontconfig
