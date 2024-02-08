test:
	go test -race -v ./...

generate-openapi:
	oapi-codegen -package jupiter -generate client,types ./jupiter/openapi/jupiter-swagger.yaml > ./jupiter/client.gen.go

lint-fix:
	golangci-lint run -E gofumpt --fix ./...

lint:
	golangci-lint run -E gofumpt ./...
