.PHONY: all build test clean lint

BINARY_NAME=ice

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/ice/main.go

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

lint:
	golangci-lint run
