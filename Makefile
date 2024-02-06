test:
	go test -race -v ./...

generate-openapi:
	oapi-codegen -package openapi -generate client,types ./openapi/jupiter-swagger.yaml > ./openapi/client.gen.go

lint-fix:
	golangci-lint run -E gofumpt --fix ./...

lint:
	golangci-lint run -E gofumpt ./...