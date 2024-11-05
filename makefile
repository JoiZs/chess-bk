
fmt:
	@go fmt ./...

vet:fmt
	@go vet ./...

dev:
	@go run main.go

build:vet
	@go build -o bin/main

test:
	@go test ./tests/ -v
