#!/bin/bash
# quick-start.sh - Quick start development environment

echo "🚀 Phoenix Platform Quick Start"
echo ""

# Check Docker
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check Go
if ! go version > /dev/null 2>&1; then
    echo "❌ Go is not installed. Please install Go 1.21+"
    exit 1
fi

echo "✅ Prerequisites checked"
echo ""

# Setup local environment
echo "Setting up local development environment..."
./scripts/setup-dev-env.sh

echo ""
echo "✅ Development environment ready!"
echo ""
echo "Available commands:"
echo "  make dev        - Start all services locally"
echo "  make test       - Run all tests"
echo "  make build      - Build all services"
echo "  make validate   - Run validation checks"
echo ""
echo "To deploy to Kubernetes:"
echo "  ./scripts/deploy-dev.sh"
