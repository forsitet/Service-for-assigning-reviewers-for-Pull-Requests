# Определяем, где лежат бинарники Go-инструментов
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: init gen-oapi lint format check-format

init:
	@echo "Initializing project tools..."
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	go install mvdan.cc/gofumpt@v0.6.0
	go install github.com/daixiang0/gci@v0.13.7
	go mod tidy

gen-oapi: init
	@echo "Generating Go code from OpenAPI specification..."
	$(GOBIN)/oapi-codegen \
		-generate types,server \
		-o ./api/openapi/openapi.gen.go \
		-package openapi \
		./api/openapi/openapi.yml

lint: init check-format
	@echo "Linter"
	$(GOBIN)/golangci-lint run ./...

format: init
	@echo "Formatting code with gofumpt..."
	$(GOBIN)/gofumpt -w .

	@echo "Formatting imports with gci..."
	$(GOBIN)/gci write --skip-generated -s Standard -s Default -s "Prefix(people)" .

check-format: init
	@echo "Checking formatting..."
	@changed="$$( $(GOBIN)/gofumpt -l . )"; \
	if [ -n "$$changed" ]; then \
		echo "Code is not formatted with gofumpt. Run 'make format' to fix."; \
		echo "$$changed"; \
		exit 1; \
	fi
