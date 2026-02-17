.PHONY: run test fmt tidy build

run:
	go run ./cmd/game

test:
	go test ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

build:
	go build -o bin/game ./cmd/game