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
test-all:
	go test ./internal/...

test-all-verbose:
	go test -v ./internal/...

.PHONY: docker-build docker-up docker-up-headless docker-down docker-logs docker-restart docker-shell
docker-build:
	docker-compose build

docker-up:
	docker-compose up

docker-up-headless:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart

docker-shell:
	docker-compose exec queue-svc sh

.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  proto    - Generate Go code from .proto files"
	@echo "  test-all - Run all tests"
	@echo "  test-all-verbose - Run all tests with verbose output"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start Docker containers"
	@echo "  docker-up-headless - Start Docker containers in detached mode"
	@echo "  docker-down    - Stop Docker containers"
	@echo "  docker-logs    - View logs from Docker containers"
	@echo "  docker-restart - Restart Docker containers"
	@echo "  docker-shell   - Open a shell in the queue-svc container"
