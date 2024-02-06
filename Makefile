generate-openapi:
	oapi-codegen -package openapi -generate client,types ./openapi/jupiter-swagger.yaml > ./openapi/client.gen.go
