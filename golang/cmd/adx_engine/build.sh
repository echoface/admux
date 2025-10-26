#!/bin/bash

# ADMUX ADX Engine Build Script
# This script compiles the adx_engine binary and packages it with configuration files

set -e  # Exit on any error

echo "Building ADMUX ADX Engine..."

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/dist/adx_engine"
BINARY_NAME="adx_engine"

# Clean previous build
echo " Cleaning previous build..."
rm -rf "${BUILD_DIR}"

# Create build directory
echo "Creating build directory..."
mkdir -p "${BUILD_DIR}"

# Build the binary
echo "Compiling adx_engine..."
cd "${PROJECT_ROOT}"
go build -o "${BUILD_DIR}/${BINARY_NAME}" ./cmd/adx_engine

# Copy configuration files
echo "Copying configuration files..."
cp -r "${PROJECT_ROOT}/cmd/adx_engine/conf" "${BUILD_DIR}/"

# Create README file
echo " Creating README..."
cat > "${BUILD_DIR}/README.md" << 'EOF'
# ADMUX ADX Engine

This directory contains the built ADX Engine binary and configuration files.

## Files
- `adx_engine` - The main executable binary
- `conf/` - Configuration files for different environments
- `README.md` - This file

## Usage

1. Run the binary:
   ```bash
   ./adx_engine
   ```

2. The application will look for configuration files in the `conf/` directory.

## Configuration

- `conf/test.yaml` - Test environment configuration
- `conf/prod.yaml` - Production environment configuration

Set the `RUN_TYPE` environment variable to select the configuration:
- `RUN_TYPE=test` for test configuration
- `RUN_TYPE=prod` for production configuration

EOF

# Set executable permissions
echo "=' Setting executable permissions..."
chmod +x "${BUILD_DIR}/${BINARY_NAME}"

# Display build summary
echo ""
echo " Build completed successfully!"
echo " Build artifacts are located in: ${BUILD_DIR}"
echo ""
echo " Build Summary:"
echo "   - Binary: ${BUILD_DIR}/${BINARY_NAME}"
echo "   - Configurations: ${BUILD_DIR}/conf/"
echo "   - README: ${BUILD_DIR}/README.md"
echo ""
echo " To run the application:"
echo "   cd ${BUILD_DIR}"
echo "   ./${BINARY_NAME}"
