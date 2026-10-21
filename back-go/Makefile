.PHONY: run build test fmt vet

run:
	go run .

build:
	go build -o bin/server .

test:
	go test ./...

fmt:
	gofmt -w main.go internal

vet:
	go vet ./...
