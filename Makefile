.PHONY: generate vet

all: server

run: server
	./server

server: go.mod go.sum $(wildcard *.go **/*.go)
	go build -o server

generate: $(wildcard *.go **/*.go)
	go generate ./...

vet: generate
	go vet ./...

test: vet generate
	go test ./...

coverage.txt: vet generate go.mod go.sum
	go test ./... -race -coverprofile=$@ -covermode=atomic

clean:
	rm -f server
