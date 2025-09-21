.PHONY: build build-adx build-tracking run-adx run-tracking clean test proto

# Build both services
build: proto build-adx build-tracking

# Build ADX server
build-adx:
	@echo "Building ADX server..."
	go build -o bin/adx_server ./cmd/adx_server

# Build Tracking server
build-tracking:
	@echo "Building Tracking server..."
	go build -o bin/trcking_server ./cmd/trcking_server

# Run ADX server
run-adx: build-adx
	@echo "Starting ADX server..."
	./bin/adx_server

# Run Tracking server
run-tracking: build-tracking
	@echo "Starting Tracking server..."
	./bin/trcking_server

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf api/gen/go/

#--go_opt=paths=source_relative \
# Generate Go code from protobuf files
proto:
	@echo "Generating Go code from protobuf..."
	mkdir -p api/gen
	PATH="$(shell go env GOPATH)/bin:$$PATH" protoc --proto_path=. \
		--go_out=api/gen \
		api/idl/*.proto
