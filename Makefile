init:
	echo "Initializing project dependencies..."
	go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
	go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	go install tool
	go mod tidy

gen-oapi:
	make init
	echo "Generating Go code from OpenAPI specification..."
	oapi-codegen \
	-generate types,server \
	-o ./api/openapi/openapi.gen.go \
	-package openapi \
	./api/openapi/openapi.yml

lint: check-format
	echo "Linter"
	go tool golangci-lint run ./...

format:
	echo "Formatting code..."
	go tool golangci-lint fmt

check-format: 
	echo "Checking formatting..."
	go tool golangci-lint fmt --diff > /dev/null || \
	(echo "Found unforamtted files. Run 'make format' to fix them"; exit 1)
