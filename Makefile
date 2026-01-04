PROTO_DIR := proto
GO_MODULE := github.com/turnertastic1/boltq

PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

.PHONY: proto
proto:
	protoc \
	  --go_out=. --go_opt=module=$(GO_MODULE) \
	  --go-grpc_out=. --go-grpc_opt=module=$(GO_MODULE) \
	  $(PROTO_FILES)
