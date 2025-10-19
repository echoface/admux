# ADMUX Project Makefile
.PHONY: proto clean-proto help

# Variables
PROTO_DIR = shared/proto
PROTO_OUT_DIR = services/admux_common/protogen
PROTO_FILES = $(shell find $(PROTO_DIR) -name "*.proto")

# Default target
all: proto

# Generate Go code from all proto files
proto:
	@echo "🔨 Generating protobuf code..."
	@mkdir -p $(PROTO_OUT_DIR)
	@export GOBIN=$$(go env GOPATH)/bin && export PATH=$$GOBIN:$$PATH && \
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
	protoc --go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		-I$(PROTO_DIR) $(PROTO_FILES)
	@echo "✅ Generated $$(find $(PROTO_OUT_DIR) -name '*.go' | wc -l) Go files"

# Clean generated proto files
clean-proto:
	@rm -rf $(PROTO_OUT_DIR)
	@echo "✅ Cleaned proto files"

# Help
help:
	@echo "Commands:"
	@echo "  proto       - Generate Go code from all .proto files"
	@echo "  clean-proto - Remove generated files"
	@echo "  help        - Show this help"
