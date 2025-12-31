NAME = lazyworktree

all: build

mkdir:
	mkdir -p bin

build: mkdir
	go build -o bin/$(NAME) ./cmd/$(NAME)

sanity: lint format test

lint:
	golangci-lint run --fix ./...

format:
	gofumpt -w .

test:
	go test -v ./...

.PHONY: all build lint format test sanity mkdir
