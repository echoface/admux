# ADX System Makefile
.PHONY: build test clean deps run proto proto-check clean-proto help

# Variables
PROTO_DIR = ../../shared/proto
KUAISHOU_PROTO = $(PROTO_DIR)/kuaishou/rtb.proto
PROTO_OUT_DIR = ../admux_common/protogen
PROTO_PACKAGE = github.com/echoface/admux/services/admux_common/protogen

# Default target
all: build

# Proto generation (with auto-install)
proto:
	@echo "🔨 Generating protobuf code for kuaishou..."
	@mkdir -p $(PROTO_OUT_DIR)
	@export GOBIN=$$(go env GOPATH)/bin && export PATH=$$GOBIN:$$PATH && \
	which protoc-gen-go > /dev/null || (echo "📦 Installing protoc-gen-go..." && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest) && \
	which protoc-gen-go-grpc > /dev/null || (echo "📦 Installing protoc-gen-go-grpc..." && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest) && \
	protoc --go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		-I$(PROTO_DIR) $(KUAISHOU_PROTO)
	@echo "✅ Proto generation completed!"
	@echo "📁 Generated files: $$(find $(PROTO_OUT_DIR) -name '*.go' | wc -l) Go files"
	@ls -la $(PROTO_OUT_DIR)/*/

# Check proto files exist
proto-check:
	@echo "🔍 Checking proto files..."
	@test -f $(KUAISHOU_PROTO) || (echo "❌ Proto file not found: $(KUAISHOU_PROTO)" && exit 1)
	@echo "✅ Proto file found: $(KUAISHOU_PROTO)"
	@echo "📊 Proto file size: $$(wc -l < $(KUAISHOU_PROTO)) lines"

# Clean generated proto files
clean-proto:
	@echo "🗑️  Cleaning generated proto files..."
	@rm -rf $(PROTO_OUT_DIR)
	@echo "✅ Proto files cleaned!"

# Build target
build:
	@echo "Building ADX Engine..."
	go build -o bin/adx_engine ./cmd/main.go
	@echo "Build completed!"

# Test target
test:
	@echo "Running tests..."
	go test ./...

# Clean target
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f *.log
	@echo "Clean completed!"

# Dependencies target
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
	@echo "Dependencies installed!"

# Run target
run:
	@echo "Running ADX Engine..."
	go run ./cmd/main.go

# Help target
help:
	@echo "🚀 ADX Engine Makefile Commands:"
	@echo ""
	@echo "🔧 Build & Run:"
	@echo "  build        - Build the ADX Engine binary"
	@echo "  run          - Run ADX Engine in development mode"
	@echo "  test         - Run all tests"
	@echo "  deps         - Install and update Go dependencies"
	@echo ""
	@echo "📦 Protobuf:"
	@echo "  proto        - Generate Go code from .proto files"
	@echo "  proto-check  - Check if proto files exist and valid"
	@echo "  clean-proto  - Remove generated proto files"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  clean        - Clean build artifacts and binaries"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "📁 Output Locations:"
	@echo "  Binary:      ./bin/adx_engine"
	@echo "  Proto files: $(PROTO_OUT_DIR)"
