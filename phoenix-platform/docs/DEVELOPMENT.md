# Phoenix Platform Development Guide

This guide covers the development workflow, standards, and best practices for contributing to the Phoenix platform.

## Table of Contents

1. [Development Environment Setup](#development-environment-setup)
2. [Project Structure](#project-structure)
3. [Development Workflow](#development-workflow)
4. [Coding Standards](#coding-standards)
5. [Testing Guidelines](#testing-guidelines)
6. [Debugging Tips](#debugging-tips)
7. [Contributing](#contributing)

## Development Environment Setup

### Prerequisites

- Go 1.21+
- Node.js 18+ and npm
- Docker and Docker Compose
- Kubernetes 1.28+ (kind or minikube for local development)
- PostgreSQL 15+ (or use Docker)
- Git

### Initial Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/phoenix/platform.git
   cd platform/phoenix-platform
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start development services**
   ```bash
   docker-compose up -d postgres prometheus grafana
   ```

5. **Run database migrations**
   ```bash
   make migrate
   ```

## Project Structure

```
phoenix-platform/
├── cmd/                    # Application entry points
│   ├── api/               # API server
│   ├── controller/        # Experiment controller
│   ├── generator/         # Config generator
│   └── simulator/         # Process simulator
├── pkg/                   # Public packages
│   ├── api/              # API business logic
│   ├── auth/             # Authentication
│   ├── generator/        # Config generation
│   └── models/           # Data models
├── internal/              # Private packages
├── operators/             # Kubernetes operators
│   ├── pipeline/         # Pipeline operator
│   └── loadsim/          # Load simulation operator
├── dashboard/             # React frontend
├── pipelines/            # Pipeline templates
├── k8s/                  # Kubernetes manifests
├── helm/                 # Helm charts
└── docs/                 # Documentation
```

## Development Workflow

### Running Services Locally

1. **API Server**
   ```bash
   go run cmd/api/main.go
   ```

2. **Dashboard Development**
   ```bash
   cd dashboard
   npm run dev
   ```

3. **Running Tests**
   ```bash
   # All tests
   make test

   # Specific package
   go test ./pkg/api/...

   # With coverage
   go test -cover ./...
   ```

### Working with Kubernetes

1. **Local Kubernetes Cluster**
   ```bash
   make cluster-up
   ```

2. **Deploy to Local Cluster**
   ```bash
   make deploy-dev
   ```

3. **Port Forwarding**
   ```bash
   # API
   kubectl port-forward svc/phoenix-api 8080:8080

   # Dashboard
   kubectl port-forward svc/phoenix-dashboard 3000:80
   ```

### Building and Packaging

1. **Build Binaries**
   ```bash
   make build
   ```

2. **Build Docker Images**
   ```bash
   make docker
   ```

3. **Generate Code**
   ```bash
   # Generate CRDs
   make generate

   # Generate protobuf
   make proto
   ```

## Coding Standards

### Go Code

1. **Style Guide**
   - Follow [Effective Go](https://golang.org/doc/effective_go.html)
   - Use `gofmt` and `golangci-lint`
   - Package names should be lowercase, single-word

2. **Project Conventions**
   ```go
   // Package comment
   // Package api provides the core business logic for experiments.
   package api

   // Exported types need comments
   // ExperimentService handles experiment lifecycle management.
   type ExperimentService struct {
       store Store
       log   *zap.Logger
   }
   ```

3. **Error Handling**
   ```go
   // Wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to create experiment: %w", err)
   }
   ```

### TypeScript/React Code

1. **Style Guide**
   - Use TypeScript strict mode
   - Follow React hooks best practices
   - Use functional components

2. **Component Structure**
   ```typescript
   interface Props {
       experiment: Experiment;
       onUpdate: (exp: Experiment) => void;
   }

   export const ExperimentCard: React.FC<Props> = ({ experiment, onUpdate }) => {
       // Component logic
   };
   ```

### Commit Messages

Follow conventional commits:
```
feat: add experiment comparison view
fix: resolve memory leak in collector
docs: update pipeline configuration guide
chore: upgrade dependencies
```

## Testing Guidelines

### Unit Tests

1. **Go Tests**
   ```go
   func TestExperimentService_Create(t *testing.T) {
       // Arrange
       service := NewExperimentService(mockStore, logger)
       
       // Act
       exp, err := service.Create(ctx, request)
       
       // Assert
       require.NoError(t, err)
       assert.Equal(t, "test-exp", exp.Name)
   }
   ```

2. **React Tests**
   ```typescript
   describe('ExperimentCard', () => {
       it('should display experiment name', () => {
           render(<ExperimentCard experiment={mockExp} />);
           expect(screen.getByText('Test Experiment')).toBeInTheDocument();
       });
   });
   ```

### Integration Tests

```bash
# Run integration tests
make test-integration

# Run specific integration test
go test -tags=integration ./test/integration/api_test.go
```

### E2E Tests

```bash
# Run E2E tests
make test-e2e
```

## Debugging Tips

### API Debugging

1. **Enable Debug Logging**
   ```bash
   LOG_LEVEL=debug go run cmd/api/main.go
   ```

2. **Use Delve Debugger**
   ```bash
   dlv debug cmd/api/main.go
   ```

3. **Inspect gRPC Calls**
   ```bash
   grpcurl -plaintext localhost:5050 list
   ```

### Kubernetes Debugging

1. **View Pod Logs**
   ```bash
   kubectl logs -f deployment/phoenix-api
   ```

2. **Exec into Pod**
   ```bash
   kubectl exec -it deployment/phoenix-api -- /bin/sh
   ```

3. **Describe Resources**
   ```bash
   kubectl describe phoenixexperiment my-experiment
   ```

### Dashboard Debugging

1. **React Developer Tools**
   - Install browser extension
   - Inspect component props and state

2. **Network Tab**
   - Monitor API calls
   - Check request/response payloads

## Contributing

### Pre-commit Checks

1. **Run Linters**
   ```bash
   make lint
   ```

2. **Format Code**
   ```bash
   make fmt
   ```

3. **Run Tests**
   ```bash
   make test
   ```

### Pull Request Process

1. Create feature branch from `main`
2. Make changes following coding standards
3. Add/update tests
4. Update documentation
5. Run `make pre-commit`
6. Submit PR with clear description

### Code Review Checklist

- [ ] Tests pass
- [ ] Code follows style guide
- [ ] Documentation updated
- [ ] No security vulnerabilities
- [ ] Performance impact considered
- [ ] Backwards compatibility maintained

## Additional Resources

- [Architecture Overview](architecture.md)
- [API Reference](api-reference.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Phoenix Slack Channel](#phoenix-dev)