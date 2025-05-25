#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Generating Protocol Buffer files...${NC}"

# Ensure protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}protoc is not installed. Please install protocol buffer compiler.${NC}"
    exit 1
fi

# Base directories
PROTO_DIR="packages/contracts/proto"
OUT_DIR="packages/contracts/proto"

# Create output directories
mkdir -p "${OUT_DIR}/v1"

# Generate Go code for proto files
for proto in "${PROTO_DIR}/v1"/*.proto; do
    if [ -f "$proto" ]; then
        echo "Generating Go code for $(basename "$proto")"
        protoc \
            --go_out="${OUT_DIR}" \
            --go_opt=paths=source_relative \
            --go-grpc_out="${OUT_DIR}" \
            --go-grpc_opt=paths=source_relative \
            -I "${PROTO_DIR}" \
            "$proto"
    fi
done

echo -e "${GREEN}Protocol Buffer generation complete!${NC}"