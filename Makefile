BINARY_NAME=bin/

all: build
 
build:
	go build -o $(BINARY_NAME) ./cmd/...

run:
	go build -o $(BINARY_NAME) ./cmd/...
	./$(BINARY_NAME)/store

lint:
	golangci-lint run
 
test:
	CGO_ENABLED=1 go test kvstore/... -race
