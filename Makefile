.PHONY: test lint fmt vet ci

test:
	go test ./...

lint: vet
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, running go vet only"; \
	fi

fmt:
	gofmt -l -w .

vet:
	go vet ./...

ci: fmt vet test
