PROTO_DIR := proto
GO_MODULE := github.com/turnertastic1/boltq

PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

.PHONY: proto
proto:
	protoc \
	  --go_out=. --go_opt=module=$(GO_MODULE) \
	  --go-grpc_out=. --go-grpc_opt=module=$(GO_MODULE) \
	  $(PROTO_FILES)

.PHONY: test-all test-all-verbose
# Run all tests
test-all:
	go test ./...

# Run all tests with verbose output
test-all-verbose:
	go test -v ./...

.PHONY: run
run:
	go run cmd/queue-svc/main.go

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  proto    - Generate Go code from .proto files"
	@echo "  test-all - Run all tests"
	@echo "  test-all-verbose - Run all tests with verbose output"
	@echo "  run      - Run the queue service"