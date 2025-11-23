init:
	echo "Initializing project dependencies..."
	go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
	go install tool
	go mod tidy

gen-oapi:
	echo "Generating Go code from OpenAPI specification..."
	oapi-codegen \
	-generate types,server \
	-o ./api/openapi/openapi.gen.go \
	-package openapi \
	./api/openapi/openapi.yml
