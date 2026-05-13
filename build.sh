#!/usr/bin/env bash

# PoH Blockchain - Build All Binaries Script
# Compiles all project binaries with proper error handling

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BIN_DIR="bin"
BUILD_FLAGS="-v"
LDFLAGS="-s -w"  # Strip debug info for smaller binaries

# Print colored message
print_msg() {
    local color=$1
    local msg=$2
    echo -e "${color}${msg}${NC}"
}

# Print section header
print_header() {
    echo ""
    print_msg "$BLUE" "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    print_msg "$BLUE" "  $1"
    print_msg "$BLUE" "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Build a single binary
build_binary() {
    local name=$1
    local source=$2
    local output="${BIN_DIR}/${name}"
    
    print_msg "$YELLOW" "Building ${name}..."
    
    if go build ${BUILD_FLAGS} -ldflags="${LDFLAGS}" -o "${output}" "${source}"; then
        local size=$(du -h "${output}" | cut -f1)
        print_msg "$GREEN" "✓ ${name} built successfully (${size})"
        return 0
    else
        print_msg "$RED" "✗ Failed to build ${name}"
        return 1
    fi
}

# Main build process
main() {
    print_header "PoH Blockchain - Build All Binaries"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_msg "$RED" "Error: Go is not installed or not in PATH"
        exit 1
    fi
    
    print_msg "$BLUE" "Go version: $(go version)"
    
    # Create bin directory if it doesn't exist
    if [ ! -d "${BIN_DIR}" ]; then
        print_msg "$YELLOW" "Creating ${BIN_DIR} directory..."
        mkdir -p "${BIN_DIR}"
    fi
    
    # Download dependencies
    print_header "Downloading Dependencies"
    print_msg "$YELLOW" "Running go mod download..."
    if go mod download; then
        print_msg "$GREEN" "✓ Dependencies downloaded"
    else
        print_msg "$RED" "✗ Failed to download dependencies"
        exit 1
    fi
    
    # Build all binaries
    print_header "Building Binaries"
    
    local failed=0
    
    # Audit binary
    build_binary "audit" "./cmd/audit/main.go" || ((failed++))
    
    # Wallet binary
    build_binary "neon-wallet" "./cmd/wallet/main.go" || ((failed++))
    
    # Validator binary
    build_binary "validator" "./cmd/validator/main.go" || ((failed++))
    
    # Devnet binary
    build_binary "devnet" "./cmd/devnet/main.go" || ((failed++))
    
    # Build binary
    build_binary "build" "./cmd/build/main.go" || ((failed++))
    
    # Summary
    print_header "Build Summary"
    
    if [ $failed -eq 0 ]; then
        print_msg "$GREEN" "✓ All binaries built successfully!"
        echo ""
        print_msg "$BLUE" "Built binaries:"
        ls -lh "${BIN_DIR}"
        echo ""
        print_msg "$BLUE" "Usage:"
        print_msg "$NC" "  ${BIN_DIR}/audit --help"
        print_msg "$NC" "  ${BIN_DIR}/neon-wallet --help"
        print_msg "$NC" "  ${BIN_DIR}/validator --help"
        print_msg "$NC" "  ${BIN_DIR}/devnet --help"
        print_msg "$NC" "  ${BIN_DIR}/build --help"
        exit 0
    else
        print_msg "$RED" "✗ ${failed} binary(ies) failed to build"
        exit 1
    fi
}

# Run main function
main "$@"
