.PHONY: run test fmt tidy build

run:
	go run ./cmd/hello

test:
	go test ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

build:
	go build -o bin/hello ./cmd/hello