.PHONY: run run-editor test fmt tidy build build-editor build-all

run:
	go run ./cmd/game

run-editor:
	go run ./cmd/editor

test:
	go test ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

build:
	go build -o bin/game ./cmd/game

build-editor:
	go build -o bin/editor ./cmd/editor

build-all: build build-editor
