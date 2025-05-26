# Protobuf Setup Guide for Phoenix Platform

## Overview

This guide explains how to set up Protocol Buffers (protobuf) generation for the Phoenix Platform's gRPC services.

## Prerequisites

### 1. Install protoc (Protocol Buffers Compiler)

#### macOS
```bash
brew install protobuf
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install -y protobuf-compiler
```

#### Verify Installation
```bash
protoc --version
# Should output: libprotoc 3.x.x
```

### 2. Install Go Protocol Buffers Plugins

```bash
# Install the protocol compiler plugins for Go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Update PATH to include Go bin directory
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Proto Files Location

Proto files are located in:
```
packages/contracts/proto/
├── v1/
│   ├── common.proto
│   ├── controller.proto
│   ├── experiment.proto
│   └── generator.proto
└── phoenix/
    └── v1/
        ├── controller.proto
        ├── experiment.proto
        └── generator.proto
```

## Generation Process

### 1. Use the Generation Script

We've created a generation script at `packages/contracts/generate.sh`:

```bash
cd packages/contracts
chmod +x generate.sh
./generate.sh
```

### 2. Manual Generation

If you need to generate manually:

```bash
cd packages/contracts

# Generate Go code for each proto file
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    -I . \
    proto/v1/common.proto

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    -I . \
    proto/v1/experiment.proto

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    -I . \
    proto/v1/controller.proto

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    -I . \
    proto/v1/generator.proto
```

## Generated Files

After running the generation, you'll have:
- `*.pb.go` - Protocol buffer message definitions
- `*_grpc.pb.go` - gRPC service definitions

Example:
```
proto/v1/
├── common.proto
├── common.pb.go           # Generated
├── controller.proto
├── controller.pb.go       # Generated
├── controller_grpc.pb.go  # Generated
├── experiment.proto
├── experiment.pb.go       # Generated
├── experiment_grpc.pb.go  # Generated
├── generator.proto
├── generator.pb.go        # Generated
└── generator_grpc.pb.go   # Generated
```

## Re-enabling Proto Code

After generating the proto files, you can re-enable the commented proto code:

### 1. Controller Service

In `projects/controller/internal/grpc/simple_server.go`:
- Uncomment the proto import
- Uncomment the `pb.UnimplementedExperimentServiceServer` embedded field
- Uncomment all gRPC handler methods

In `projects/controller/internal/grpc/simple_server_test.go`:
- Uncomment the proto import
- Uncomment test functions that use proto types

In `projects/controller/cmd/controller/main.go`:
- Uncomment the proto import
- Uncomment the service registration: `pb.RegisterExperimentServiceServer(grpcServer, adapterServer)`

### 2. Platform API Service

If you add gRPC support to platform-api:
- Import the generated proto packages
- Implement the gRPC service interfaces
- Register services with the gRPC server

### 3. Generator Service

Similar process for the generator service to use the generated proto definitions.

## CI/CD Integration

Add proto generation to your CI/CD pipeline:

```yaml
# .github/workflows/proto-check.yml
name: Proto Check

on:
  pull_request:
    paths:
      - '**.proto'

jobs:
  proto-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install protoc
        run: |
          sudo apt-get update
          sudo apt-get install -y protobuf-compiler
          
      - name: Install Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Install protoc plugins
        run: |
          go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
          go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
          
      - name: Generate proto
        run: |
          cd packages/contracts
          ./generate.sh
          
      - name: Check for uncommitted changes
        run: |
          if [[ -n $(git status -s) ]]; then
            echo "Proto files need to be regenerated"
            exit 1
          fi
```

## Troubleshooting

### "protoc-gen-go: program not found"
Make sure Go bin directory is in your PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Import errors after generation
Ensure the `go_package` option in proto files matches your module structure:
```protobuf
option go_package = "github.com/phoenix-vnext/platform/packages/contracts/proto/v1;v1";
```

### Missing dependencies
Run `go mod tidy` in the contracts package after generation:
```bash
cd packages/contracts
go mod tidy
```

## Next Steps

1. Install protoc and plugins
2. Generate proto files
3. Re-enable commented proto code in services
4. Test gRPC endpoints
5. Add proto generation to CI/CD pipeline

The Phoenix Platform will then have fully functional gRPC services with type-safe protocol buffer definitions!