# Development tasks for jupiter-go. Run `make` or `make help` to list targets.
# Targets avoid features newer than GNU Make 3.81, the version macOS ships.

GO            ?= go
GOLANGCI_LINT ?= golangci-lint
PKGS          ?= ./...

# FUZZTIME bounds each fuzz target; COVERPROFILE is the coverage output file.
FUZZTIME     ?= 30s
COVERPROFILE ?= coverage.out

# Fuzz targets as package:function pairs. Go runs one -fuzz target per
# invocation, so `make fuzz` runs them in turn.
FUZZ_TARGETS := \
	./solanatx:FuzzParse \
	./internal/address:FuzzValidate \
	./internal/amount:FuzzParseRaw \
	./lend:FuzzParseSignedAmount

.DEFAULT_GOAL := help

.PHONY: help check test vet lint lint-fix fmt tidy cover fuzz clean

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check: vet lint tidy test ## Run everything CI runs, plus the go.mod tidiness check.

test: ## Run tests with the race detector.
	$(GO) test -race $(PKGS)

vet: ## Run go vet.
	$(GO) vet $(PKGS)

# golangci-lint v2 applies gofumpt as a formatter, through the fmt subcommand.
lint: ## Run golangci-lint and report gofumpt differences.
	$(GOLANGCI_LINT) run $(PKGS)
	$(GOLANGCI_LINT) fmt -E gofumpt -d $(PKGS)

lint-fix: ## Apply golangci-lint fixes and gofumpt formatting.
	$(GOLANGCI_LINT) run --fix $(PKGS)
	$(GOLANGCI_LINT) fmt -E gofumpt $(PKGS)

fmt: ## Format code with gofumpt.
	$(GOLANGCI_LINT) fmt -E gofumpt $(PKGS)

tidy: ## Fail if go.mod or go.sum is not tidy.
	$(GO) mod tidy -diff

cover: ## Run tests with coverage and print the per-function summary.
	$(GO) test -race -covermode=atomic -coverpkg=$(PKGS) -coverprofile=$(COVERPROFILE) $(PKGS)
	$(GO) tool cover -func=$(COVERPROFILE)

fuzz: ## Run each fuzz target for FUZZTIME (default 30s).
	@set -e; for target in $(FUZZ_TARGETS); do \
		pkg=$${target%%:*}; \
		name=$${target#*:}; \
		echo "fuzzing $$name in $$pkg for $(FUZZTIME)"; \
		$(GO) test -run='^$$' -fuzz="^$$name\$$" -fuzztime=$(FUZZTIME) $$pkg; \
	done

clean: ## Remove coverage output and the test cache.
	rm -f $(COVERPROFILE)
	$(GO) clean -testcache
