.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY: fmt

lint: fmt
	staticcheck
.PHONY: lint

vet: fmt
	go vet ./...
.PHONY: vet

build: vet
	go mod tidy
	go build -ldflags="-s -w" -o build/ringring cmd/ringring/main.go
	go build -ldflags="-s -w" -o build/deploy cmd/deploy/main.go
.PHONY: build