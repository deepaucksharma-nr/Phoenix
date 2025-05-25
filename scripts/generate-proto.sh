#!/bin/bash

# Generate Proto Files Script
# This script generates Go code from protobuf definitions

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Generating Proto Files...${NC}"

# Change to pkg directory
cd "${PROJECT_ROOT}/pkg"

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

# Create output directories
mkdir -p grpc/proto/v1/{common,experiment,controller,generator}

# Generate proto files
echo -e "${GREEN}Generating common proto...${NC}"
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    -I grpc/proto \
    grpc/proto/v1/common.proto

echo -e "${GREEN}Generating experiment proto...${NC}"
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    -I grpc/proto \
    grpc/proto/v1/experiment.proto

echo -e "${GREEN}Generating controller proto...${NC}"
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    -I grpc/proto \
    grpc/proto/v1/controller.proto

echo -e "${GREEN}Generating generator proto...${NC}"
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    -I grpc/proto \
    grpc/proto/v1/generator.proto

echo -e "${GREEN}Proto generation complete!${NC}"

# List generated files
echo -e "${YELLOW}Generated files:${NC}"
find grpc/proto/v1 -name "*.pb.go" -o -name "*_grpc.pb.go" | sort