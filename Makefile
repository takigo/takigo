# Convenience wrappers around the commands in AGENTS.md. Run `make help`.

GO     ?= go
PKGS   ?= ./...
DEMO   ?= button
OUTDIR ?= tmp/bin

.DEFAULT_GOAL := help

.PHONY: help all build test test-race vet fmt fmt-check fix fix-check lint \
	check ci build-windows vet-windows demos run-demo bench fuzz clean distclean

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

fmt: ## Format all Go files
	gofmt -w .

fmt-check: ## Fail if any file needs gofmt
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "needs gofmt:"; echo "$$out"; exit 1; fi

fix: ## Apply the go fix modernizers
	$(GO) fix $(PKGS)

fix-check: ## Fail if go fix would change anything (CI gate)
	$(GO) fix -diff $(PKGS)

lint: ## golangci-lint (CI only gates changed lines)
	golangci-lint run $(PKGS)

check: vet fmt-check fix-check vet-windows ## The static checks CI gates on

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

clean: ## Remove built binaries
	rm -rf $(OUTDIR)/demos $(OUTDIR)/windows

distclean: clean ## clean, plus Go build/test/fuzz caches and generated tmp/ files
	$(GO) clean -cache -testcache -fuzzcache
	rm -rf $(OUTDIR) tmp/fontconfig
