
fmt:
	@go fmt ./...

vet: fmt
	@go vet ./...

build: vet
	@go build -o bin/main

dev:
	@go run main.go
