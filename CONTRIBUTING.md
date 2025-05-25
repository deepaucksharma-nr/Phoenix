# Contributing to Phoenix Platform

Thank you for your interest in contributing to Phoenix Platform! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md). We are committed to providing a welcoming and inclusive environment for all contributors.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Node.js 18 or later
- Docker and Docker Compose
- Make
- Git

### Setting Up Your Development Environment

1. **Fork the repository**
   ```bash
   # Click the "Fork" button on GitHub
   ```

2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/phoenix-platform.git
   cd phoenix-platform
   ```

3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/phoenix/platform.git
   ```

4. **Set up the development environment**
   ```bash
   make setup
   ```

5. **Start local services**
   ```bash
   make dev-up
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
# Update your local main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

- Write clean, readable code
- Follow the coding standards
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
make test

# Run specific project tests
make test-platform-api

# Run linting
make lint

# Run security checks
make security
```

### 4. Commit Your Changes

Follow our [commit guidelines](#commit-guidelines) for commit messages.

```bash
git add .
git commit -m "feat(api): add new optimization endpoint"
```

### 5. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 6. Create a Pull Request

Go to GitHub and create a pull request from your fork to the main repository.

## Coding Standards

### Go Code Style

We follow standard Go conventions with some additional guidelines:

```go
// Package comment should be present
package example

import (
    "context"
    "fmt"
    
    "github.com/phoenix/platform/pkg/errors"
)

// ExampleService provides example functionality.
type ExampleService struct {
    logger logging.Logger
    db     *sql.DB
}

// NewExampleService creates a new example service.
func NewExampleService(logger logging.Logger, db *sql.DB) *ExampleService {
    return &ExampleService{
        logger: logger,
        db:     db,
    }
}

// ProcessData processes the given data according to business rules.
func (s *ExampleService) ProcessData(ctx context.Context, data string) error {
    // Use structured logging
    s.logger.Info("processing data", 
        logging.String("data_length", len(data)))
    
    // Always handle errors explicitly
    if err := s.validateData(data); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // More processing...
    return nil
}
```

### TypeScript/React Code Style

```typescript
import React, { useState, useEffect } from 'react';
import { useAppDispatch, useAppSelector } from '@/hooks';

interface ExampleProps {
  title: string;
  onAction: (value: string) => void;
}

/**
 * ExampleComponent demonstrates our coding standards.
 */
export const ExampleComponent: React.FC<ExampleProps> = ({ 
  title, 
  onAction 
}) => {
  const [value, setValue] = useState('');
  const dispatch = useAppDispatch();
  const { loading, error } = useAppSelector(state => state.example);

  useEffect(() => {
    // Effect logic here
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onAction(value);
  };

  return (
    <div className="example-component">
      <h2>{title}</h2>
      <form onSubmit={handleSubmit}>
        <input
          value={value}
          onChange={(e) => setValue(e.target.value)}
          disabled={loading}
        />
        <button type="submit" disabled={loading}>
          Submit
        </button>
      </form>
      {error && <div className="error">{error}</div>}
    </div>
  );
};
```

### General Guidelines

1. **Naming Conventions**
   - Go: Use camelCase for variables, PascalCase for types
   - TypeScript: Use camelCase for variables/functions, PascalCase for types/components
   - Files: Use snake_case for Go files, kebab-case for TypeScript files

2. **Error Handling**
   - Always handle errors explicitly
   - Wrap errors with context
   - Use structured logging for errors

3. **Testing**
   - Write unit tests for all new code
   - Aim for >80% code coverage
   - Use table-driven tests in Go
   - Use React Testing Library for React components

4. **Documentation**
   - Document all exported types and functions
   - Include examples for complex functionality
   - Keep documentation up to date

## Commit Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test additions or modifications
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Other changes

### Examples

```bash
# Feature
feat(api): add experiment validation endpoint

# Bug fix
fix(collector): resolve memory leak in metric processing

# Documentation
docs(readme): update installation instructions

# With scope and breaking change
feat(auth)!: change JWT token format

BREAKING CHANGE: JWT tokens now include namespace claim
```

## Pull Request Process

### Before Creating a PR

1. **Ensure all tests pass**
   ```bash
   make test
   ```

2. **Run linting**
   ```bash
   make lint
   ```

3. **Update documentation**
   - Update README if needed
   - Add/update API documentation
   - Update architectural diagrams if applicable

### PR Requirements

1. **Title**: Use conventional commit format
2. **Description**: Fill out the PR template completely
3. **Tests**: All tests must pass
4. **Reviews**: Requires at least 2 approvals
5. **No conflicts**: Resolve merge conflicts

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] No new warnings
```

## Testing Requirements

### Unit Tests

Every new function should have corresponding unit tests:

```go
func TestExampleService_ProcessData(t *testing.T) {
    tests := []struct {
        name    string
        data    string
        wantErr bool
    }{
        {
            name:    "valid data",
            data:    "test data",
            wantErr: false,
        },
        {
            name:    "empty data",
            data:    "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := NewExampleService(testLogger, testDB)
            err := s.ProcessData(context.Background(), tt.data)
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessData() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

Integration tests should cover service interactions:

```go
func TestExperimentLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Setup
    client := setupTestClient(t)
    defer cleanupTestData(t)
    
    // Test experiment lifecycle
    experiment := createTestExperiment(t, client)
    startExperiment(t, client, experiment.ID)
    waitForRunning(t, client, experiment.ID)
    stopExperiment(t, client, experiment.ID)
}
```

### E2E Tests

End-to-end tests should cover complete user workflows:

```typescript
test('complete experiment workflow', async ({ page }) => {
  // Login
  await loginAs(page, 'test@example.com');
  
  // Create experiment
  await page.goto('/experiments/new');
  await fillExperimentForm(page, testExperiment);
  await page.click('button[type="submit"]');
  
  // Verify creation
  await expect(page).toHaveURL(/\/experiments\/\d+/);
  await expect(page.locator('h1')).toContainText(testExperiment.name);
});
```

## Documentation

### Code Documentation

- Document all exported types and functions
- Include examples for complex APIs
- Keep comments concise and valuable

### Project Documentation

When adding new features:

1. Update relevant README files
2. Add API documentation
3. Update architecture diagrams if needed
4. Add runbooks for operational procedures

### Documentation Style

- Use clear, concise language
- Include code examples
- Add diagrams for complex concepts
- Keep documentation close to code

## Community

### Getting Help

- **Discord**: [Join our Discord](https://discord.gg/phoenix)
- **Discussions**: Use GitHub Discussions for questions
- **Issues**: Report bugs via GitHub Issues

### Communication Guidelines

- Be respectful and inclusive
- Provide context in questions
- Search existing issues before creating new ones
- Use appropriate channels for different topics

### Recognition

We value all contributions! Contributors are recognized in:

- Release notes
- Contributors file
- Project website
- Community calls

## Additional Resources

- [Architecture Guide](docs/architecture/README.md)
- [API Documentation](docs/api/README.md)
- [Development Guide](docs/guides/developer/getting-started.md)
- [Phoenix Platform Website](https://phoenix.io)

Thank you for contributing to Phoenix Platform! 🚀