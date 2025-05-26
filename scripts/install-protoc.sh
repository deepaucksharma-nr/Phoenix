#!/bin/bash

# Install Protoc Script
# This script helps install the protobuf compiler

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing Protobuf Compiler...${NC}"

# Detect OS
OS=$(uname -s)
ARCH=$(uname -m)

if [ "$OS" = "Darwin" ]; then
    # macOS
    if command -v brew &> /dev/null; then
        echo -e "${YELLOW}Installing protoc via Homebrew...${NC}"
        brew install protobuf
    else
        echo -e "${RED}Homebrew not found. Please install Homebrew first:${NC}"
        echo "/bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        exit 1
    fi
elif [ "$OS" = "Linux" ]; then
    # Linux
    if command -v apt-get &> /dev/null; then
        # Debian/Ubuntu
        echo -e "${YELLOW}Installing protoc via apt...${NC}"
        sudo apt-get update
        sudo apt-get install -y protobuf-compiler
    elif command -v yum &> /dev/null; then
        # RHEL/CentOS
        echo -e "${YELLOW}Installing protoc via yum...${NC}"
        sudo yum install -y protobuf-compiler
    else
        # Manual installation
        echo -e "${YELLOW}Installing protoc manually...${NC}"
        PROTOC_VERSION="25.1"
        if [ "$ARCH" = "x86_64" ]; then
            PROTOC_ZIP="protoc-${PROTOC_VERSION}-linux-x86_64.zip"
        else
            PROTOC_ZIP="protoc-${PROTOC_VERSION}-linux-aarch_64.zip"
        fi
        
        curl -OL "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP}"
        sudo unzip -o "$PROTOC_ZIP" -d /usr/local bin/protoc
        sudo unzip -o "$PROTOC_ZIP" -d /usr/local 'include/*'
        rm -f "$PROTOC_ZIP"
    fi
else
    echo -e "${RED}Unsupported OS: $OS${NC}"
    exit 1
fi

# Install Go protoc plugins
echo -e "${YELLOW}Installing Go protoc plugins...${NC}"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify installation
if command -v protoc &> /dev/null; then
    echo -e "${GREEN}✓ protoc installed successfully!${NC}"
    protoc --version
else
    echo -e "${RED}✗ protoc installation failed${NC}"
    exit 1
fi

if command -v protoc-gen-go &> /dev/null; then
    echo -e "${GREEN}✓ protoc-gen-go installed successfully!${NC}"
else
    echo -e "${RED}✗ protoc-gen-go installation failed${NC}"
    exit 1
fi

if command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${GREEN}✓ protoc-gen-go-grpc installed successfully!${NC}"
else
    echo -e "${RED}✗ protoc-gen-go-grpc installation failed${NC}"
    exit 1
fi

echo -e "${GREEN}Installation complete! You can now run:${NC}"
echo "cd packages/contracts && bash generate.sh"