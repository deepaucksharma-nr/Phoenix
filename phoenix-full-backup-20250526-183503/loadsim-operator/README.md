# loadsim-operator

## Overview

[Brief description of the service]

## Architecture

[Service architecture and design decisions]

## Development

### Prerequisites

- Go 1.21+ (for Go services)
- Node.js 18+ (for Node services)
- Docker
- Make

### Setup

```bash
# Install dependencies
make install

# Run tests
make test

# Build the service
make build
```

### Running Locally

```bash
# Start development server
make dev

# Or run the built binary
make run
```

## Configuration

Configuration is managed through environment variables and config files.

See `configs/` directory for configuration examples.

## API Documentation

[Link to API documentation]

## Testing

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run with coverage
make test-coverage
```

## Deployment

```bash
# Build Docker image
make docker

# Push to registry
make docker-push
```

## Monitoring

- Metrics: Available at `/metrics`
- Health: Available at `/health`
- Ready: Available at `/ready`

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md)
