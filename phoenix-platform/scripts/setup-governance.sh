#!/bin/bash
# Setup script for Phoenix Platform governance and code quality tools

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "🚀 Phoenix Platform Governance Setup"
echo "===================================="

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    echo "Please install Go 1.21 or later: https://golang.org/dl/"
    exit 1
else
    echo -e "${GREEN}✓ Go $(go version | awk '{print $3}')${NC}"
fi

# Check Node.js
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js is not installed${NC}"
    echo "Please install Node.js 18 or later: https://nodejs.org/"
    exit 1
else
    echo -e "${GREEN}✓ Node.js $(node --version)${NC}"
fi

# Check Python (for pre-commit)
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}❌ Python 3 is not installed${NC}"
    echo "Please install Python 3: https://www.python.org/"
    exit 1
else
    echo -e "${GREEN}✓ Python $(python3 --version)${NC}"
fi

# Install Go tools
echo -e "\n${YELLOW}Installing Go tools...${NC}"

# GolangCI-Lint
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.0
    echo -e "${GREEN}✓ golangci-lint installed${NC}"
else
    echo -e "${GREEN}✓ golangci-lint already installed${NC}"
fi

# Go imports
echo "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest
echo -e "${GREEN}✓ goimports installed${NC}"

# Go fumpt (stricter gofmt)
echo "Installing gofumpt..."
go install mvdan.cc/gofumpt@latest
echo -e "${GREEN}✓ gofumpt installed${NC}"

# Install Node.js tools
echo -e "\n${YELLOW}Installing Node.js tools...${NC}"

# Commitlint
echo "Installing commitlint..."
npm install -g @commitlint/cli @commitlint/config-conventional
echo -e "${GREEN}✓ commitlint installed${NC}"

# Install Python tools
echo -e "\n${YELLOW}Installing Python tools...${NC}"

# Pre-commit
if ! command -v pre-commit &> /dev/null; then
    echo "Installing pre-commit..."
    pip3 install --user pre-commit
    echo -e "${GREEN}✓ pre-commit installed${NC}"
else
    echo -e "${GREEN}✓ pre-commit already installed${NC}"
fi

# Detect-secrets
echo "Installing detect-secrets..."
pip3 install --user detect-secrets
echo -e "${GREEN}✓ detect-secrets installed${NC}"

# Setup pre-commit hooks
echo -e "\n${YELLOW}Setting up pre-commit hooks...${NC}"
cd ..  # Go to repo root
pre-commit install
pre-commit install --hook-type commit-msg
echo -e "${GREEN}✓ Pre-commit hooks installed${NC}"

# Initialize secrets baseline
echo -e "\n${YELLOW}Initializing secrets baseline...${NC}"
detect-secrets scan --baseline .secrets.baseline
echo -e "${GREEN}✓ Secrets baseline created${NC}"

# Run initial validation
echo -e "\n${YELLOW}Running initial validation...${NC}"
cd phoenix-platform
make validate-structure || true

echo -e "\n${GREEN}✅ Governance setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Run 'make validate' to check the codebase"
echo "2. Run 'make setup-hooks' to install git hooks"
echo "3. Commit changes will now be validated automatically"
echo ""
echo "Available commands:"
echo "  make validate          - Run all validation checks"
echo "  make validate-structure - Check mono-repo structure"
echo "  make validate-imports  - Check Go import rules"
echo "  make lint             - Run linters"
echo "  make fmt              - Format code"
echo "  make test             - Run tests"
echo "  make verify           - Run all pre-commit checks"