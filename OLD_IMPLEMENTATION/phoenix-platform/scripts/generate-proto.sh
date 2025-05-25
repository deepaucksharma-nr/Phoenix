#!/bin/bash

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Directories
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PROTO_DIR="${PROJECT_ROOT}/api/proto"
OUT_DIR="${PROJECT_ROOT}/pkg/api/v1"

echo -e "${GREEN}🔧 Phoenix Platform Proto Code Generation${NC}"
echo "========================================"

# Check for protoc
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}✗ protoc not found. Please install protobuf compiler.${NC}"
    echo "  On macOS: brew install protobuf"
    echo "  On Ubuntu: apt-get install protobuf-compiler"
    exit 1
fi

# Check for Go plugins
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go-grpc...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create output directory
mkdir -p "${OUT_DIR}"

# Generate Go code from proto files
echo -e "${GREEN}Generating Go code from proto files...${NC}"

# Find all proto files
PROTO_FILES=$(find "${PROTO_DIR}" -name "*.proto")

if [ -z "$PROTO_FILES" ]; then
    echo -e "${RED}✗ No proto files found in ${PROTO_DIR}${NC}"
    exit 1
fi

# Generate Go code for each proto file
for proto_file in $PROTO_FILES; do
    echo "Processing: $(basename "$proto_file")"
    
    protoc \
        --go_out="${PROJECT_ROOT}" \
        --go_opt=paths=source_relative \
        --go-grpc_out="${PROJECT_ROOT}" \
        --go-grpc_opt=paths=source_relative \
        -I "${PROTO_DIR}" \
        -I "${PROJECT_ROOT}" \
        "$proto_file"
done

echo -e "${GREEN}✓ Proto generation complete!${NC}"

# Generate mock implementations
echo -e "${GREEN}Generating mock implementations...${NC}"
if ! command -v mockgen &> /dev/null; then
    echo -e "${YELLOW}Installing mockgen...${NC}"
    go install github.com/golang/mock/mockgen@latest
fi

# Generate mocks for each service
for proto_file in $PROTO_FILES; do
    base_name=$(basename "$proto_file" .proto)
    
    # Skip if no service is defined
    if ! grep -q "^service" "$proto_file"; then
        continue
    fi
    
    echo "Generating mocks for: ${base_name}"
    
    mockgen -source="${OUT_DIR}/${base_name}_grpc.pb.go" \
            -destination="${PROJECT_ROOT}/pkg/mocks/${base_name}_grpc_mock.go" \
            -package=mocks
done

echo -e "${GREEN}✓ Mock generation complete!${NC}"

# Update import paths in generated files
echo -e "${GREEN}Updating import paths...${NC}"
find "${OUT_DIR}" -name "*.pb.go" -exec sed -i.bak 's|github.com/phoenix-platform|github.com/phoenix/platform|g' {} \;
find "${OUT_DIR}" -name "*.pb.go.bak" -exec rm {} \;

echo -e "${GREEN}✓ All done! Generated files are in ${OUT_DIR}${NC}"