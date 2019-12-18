.PHONY: test

all: server

run: server
	./server

server: $(wildcard go.mod go.sum *.go **/*.go)
	go build -o server

test:
	go vet ./...
	go generate ./...
	go test ./...
