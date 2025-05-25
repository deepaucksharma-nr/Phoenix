# api

## Overview

Service migrated from: `phoenix-platform/cmd/api`

## Development

### Prerequisites

- Go 1.21+ (for Go services)
- Node.js 18+ (for Node services)
- Docker
- Make

### Quick Start

```bash
# Install dependencies
make install

# Run tests
make test

# Build the service
make build

# Run locally
make run   # or 'make dev' for Node services
```

## Docker

```bash
# Build Docker image
make docker

# Run with docker-compose (from root)
docker-compose up api
```

## Configuration

Configuration is managed through environment variables and config files.
See `configs/` directory for examples.

## API Documentation

[TODO: Add API documentation]

## Testing

```bash
# Unit tests
make test

# With coverage
make test-coverage
```
