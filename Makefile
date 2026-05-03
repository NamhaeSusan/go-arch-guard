.PHONY: test lint fmt vet ci

test:
	go test ./...

lint: vet
	golangci-lint run ./...

fmt:
	go list -f '{{.Dir}}' ./... | xargs -I{} find {} -maxdepth 1 -name '*.go' -print | xargs gofmt -w

vet:
	go vet ./...

ci: fmt vet test
