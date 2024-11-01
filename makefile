
fmt:
	@go fmt ./...

vet:fmt
	@go vet ./...

dev:
	@go run main.go

build:vet
	@go build -o bin/main
