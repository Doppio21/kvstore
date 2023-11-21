BINARY_NAME=bin/

PROTO_OUT=./internal/protobuf/
GOGOPROTO_DIR = $(shell go list -m -f '{{.Dir}}' github.com/gogo/protobuf)

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

proto:
	for d in ./protobuf/*; do \
		protoc --gogoslick_out=plugins=grpc,paths=source_relative:./internal/protobuf -I ./protobuf -I $(GOGOPROTO_DIR) $$d/*proto ; \
	done
