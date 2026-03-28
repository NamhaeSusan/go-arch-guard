.PHONY: test lint fmt vet ci

test:
	go test ./...

lint: vet
	golangci-lint run ./...

fmt:
	gofmt -l -w .

vet:
	go vet ./...

ci: fmt vet test
