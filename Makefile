.PHONY: test vet lint lint-fix check fuzz

test:
	go test -race ./...

vet:
	go vet ./...

# golangci-lint v2 moved gofumpt from a linter to a formatter, so it is no
# longer enabled via `run -E` and is invoked through the `fmt` subcommand.
lint:
	golangci-lint run ./...
	golangci-lint fmt -E gofumpt -d ./...

lint-fix:
	golangci-lint run --fix ./...
	golangci-lint fmt -E gofumpt ./...

# check runs everything CI runs.
check: vet lint test

# fuzz runs each fuzz target for FUZZTIME (default 30s). Go accepts one -fuzz
# target per package invocation, so each runs separately.
FUZZTIME ?= 30s

fuzz:
	go test -run=^$$ -fuzz=FuzzParse$$ -fuzztime=$(FUZZTIME) ./solanatx
	go test -run=^$$ -fuzz=FuzzValidate -fuzztime=$(FUZZTIME) ./internal/address
	go test -run=^$$ -fuzz=FuzzParseRaw -fuzztime=$(FUZZTIME) ./internal/amount
	go test -run=^$$ -fuzz=FuzzParseSignedAmount -fuzztime=$(FUZZTIME) ./lend
