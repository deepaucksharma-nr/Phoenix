# Phoenix Platform - Contract Definitions

This directory contains all API contracts and schemas that define the boundaries between services.

## Structure

```
contracts/
├── api/          # REST API OpenAPI specifications
├── grpc/         # gRPC Protocol Buffer definitions  
├── events/       # AsyncAPI event definitions
├── graphql/      # GraphQL schema definitions
└── validation/   # Contract validation tools
```

## Purpose

Contracts serve as the source of truth for:
- Service interfaces
- Data schemas
- Event formats
- API versioning

## Usage

### For Service Development

1. **Define contracts first** before implementing services
2. **Generate code** from contracts using appropriate tools
3. **Validate implementations** against contracts in CI/CD

### For LLM-based Development

Contracts provide strict boundaries that AI agents must follow:
- Generated code must conform to defined schemas
- API changes require contract updates first
- Breaking changes are detected automatically

## Contract Validation

All contracts are validated in the CI pipeline:
- OpenAPI specs are validated for correctness
- Protocol buffers are compiled and checked
- Breaking changes are detected and reported

## Version Management

- Contracts follow semantic versioning
- Breaking changes require major version bumps
- Multiple versions can coexist during migration periods