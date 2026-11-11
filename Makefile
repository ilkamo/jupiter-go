.PHONY: test vet lint lint-fix check

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
