#!/bin/bash

# Generate Proto Files for Contracts Package
# This script generates Go code from protobuf definitions

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_ROOT="${SCRIPT_DIR}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Generating Proto Files for Contracts Package...${NC}"

# Ensure protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}protoc is not installed. Please install protobuf compiler.${NC}"
    echo "On macOS: brew install protobuf"
    echo "On Ubuntu: apt-get install protobuf-compiler"
    exit 1
fi

# Ensure protoc-gen-go and protoc-gen-go-grpc are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go-grpc...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Change to package directory
cd "${PACKAGE_ROOT}"

# Generate proto files from proto/v1
echo -e "${GREEN}Generating proto files from proto/v1...${NC}"

# Generate each proto file
for proto in proto/v1/*.proto; do
    if [ -f "$proto" ]; then
        filename=$(basename "$proto")
        echo -e "${YELLOW}  Generating ${filename}...${NC}"
        protoc --go_out=. --go_opt=paths=source_relative \
            --go-grpc_out=. --go-grpc_opt=paths=source_relative \
            -I . \
            "$proto"
    fi
done

echo -e "${GREEN}Proto generation complete!${NC}"

# List generated files
echo -e "${YELLOW}Generated files:${NC}"
find proto/v1 -name "*.pb.go" -o -name "*_grpc.pb.go" | sort