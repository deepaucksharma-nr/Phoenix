# Contributing to Phoenix Platform

Thank you for your interest in contributing to Phoenix! This guide will help you get started with contributing to our project.

## 🤝 Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it before contributing.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Docker and Docker Compose
- Kubernetes cluster (kind, minikube, or cloud)
- Git

### Development Setup

1. **Fork and Clone**
   ```bash
   # Fork the repository on GitHub, then:
   git clone https://github.com/YOUR-USERNAME/Phoenix.git
   cd Phoenix
   git remote add upstream https://github.com/phoenix-platform/phoenix.git
   ```

2. **Install Dependencies**
   ```bash
   cd phoenix-platform
   make deps
   make setup-hooks
   ```

3. **Start Development Environment**
   ```bash
   make dev
   ```

## 📋 How to Contribute

### Finding Issues

- Check our [issue tracker](https://github.com/phoenix-platform/phoenix/issues)
- Look for issues labeled `good first issue` or `help wanted`
- Feel free to ask questions in issues or on Slack

### Types of Contributions

We welcome many types of contributions:

- 🐛 **Bug Fixes**: Fix bugs and improve stability
- ✨ **Features**: Add new functionality
- 📚 **Documentation**: Improve docs, add examples
- 🧪 **Tests**: Increase test coverage
- 🎨 **UI/UX**: Enhance the dashboard
- 🔧 **Refactoring**: Improve code quality

## 🔄 Development Workflow

### 1. Create a Branch

```bash
# Update your fork
git checkout main
git fetch upstream
git merge upstream/main

# Create a feature branch
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions/fixes

### 2. Make Your Changes

Follow our coding standards:

#### Go Code
```go
// Package comment
// Package experiment provides experiment management functionality.
package experiment

// Exported types need comments
// Service handles experiment lifecycle operations.
type Service struct {
    store Store
    log   *zap.Logger
}

// CreateExperiment creates a new A/B testing experiment.
func (s *Service) CreateExperiment(ctx context.Context, req *CreateRequest) (*Experiment, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Business logic here
    return experiment, nil
}
```

#### React/TypeScript Code
```typescript
// Use functional components with TypeScript
interface ExperimentCardProps {
  experiment: Experiment;
  onSelect: (id: string) => void;
}

export const ExperimentCard: React.FC<ExperimentCardProps> = ({ 
  experiment, 
  onSelect 
}) => {
  // Component logic
  return (
    <Card onClick={() => onSelect(experiment.id)}>
      <CardContent>
        <Typography variant="h6">{experiment.name}</Typography>
      </CardContent>
    </Card>
  );
};
```

### 3. Write Tests

All code changes should include tests:

#### Go Tests
```go
func TestService_CreateExperiment(t *testing.T) {
    // Arrange
    mockStore := &MockStore{}
    service := NewService(mockStore, logger)
    
    req := &CreateRequest{
        Name: "test-experiment",
        BaselinePipeline: "baseline-v1",
    }
    
    // Act
    exp, err := service.CreateExperiment(context.Background(), req)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, "test-experiment", exp.Name)
    mockStore.AssertExpectations(t)
}
```

#### React Tests
```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { ExperimentCard } from './ExperimentCard';

describe('ExperimentCard', () => {
  it('calls onSelect when clicked', () => {
    const handleSelect = vi.fn();
    const experiment = { id: '1', name: 'Test' };
    
    render(
      <ExperimentCard 
        experiment={experiment} 
        onSelect={handleSelect} 
      />
    );
    
    fireEvent.click(screen.getByText('Test'));
    expect(handleSelect).toHaveBeenCalledWith('1');
  });
});
```

### 4. Update Documentation

- Update relevant documentation in `/docs`
- Add/update API documentation if you changed APIs
- Update README if you added new features
- Add examples if applicable

### 5. Run Checks

Before committing:

```bash
# Format code
make fmt

# Run linters
make lint

# Run tests
make test

# Run all validation
make validate
```

### 6. Commit Your Changes

Follow our commit message convention:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks

Examples:
```bash
git commit -m "feat(api): add metrics export endpoint

- Add new /v1/metrics/export endpoint
- Support CSV and JSON formats
- Add rate limiting

Closes #123"
```

### 7. Push and Create PR

```bash
# Push your branch
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub with:
- Clear title and description
- Reference any related issues
- Include screenshots for UI changes
- Add test results

## 📝 Pull Request Guidelines

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests pass locally (`make test`)
- [ ] Documentation is updated
- [ ] Commit messages follow convention
- [ ] PR description explains the change
- [ ] Related issues are linked

### PR Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Screenshots (if applicable)
Add screenshots here

## Related Issues
Closes #(issue number)
```

## 🧪 Testing Guidelines

### Test Coverage

We aim for:
- 80% coverage for business logic
- 70% coverage for API handlers
- 60% coverage for UI components

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# E2E tests
make test-e2e

# Coverage report
make coverage
```

## 📚 Documentation Guidelines

### Writing Documentation

- Use clear, concise language
- Include code examples
- Add diagrams where helpful
- Keep it up to date

### Documentation Structure

```markdown
# Feature Name

## Overview
Brief description

## Usage
How to use the feature

### Example
```code
example here
```

## Configuration
Configuration options

## Troubleshooting
Common issues and solutions
```

## 🏗️ Architecture Guidelines

### Adding New Services

1. Follow the existing service structure
2. Define proto contracts first
3. Implement interfaces
4. Add to validation scripts
5. Update documentation

### Database Changes

1. Create migration files
2. Test migrations up and down
3. Update models
4. Document schema changes

## 🚢 Release Process

We use semantic versioning (MAJOR.MINOR.PATCH):

- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes

## 🆘 Getting Help

- 💬 [Slack Community](https://phoenix-community.slack.com)
- 📧 [Mailing List](https://groups.google.com/g/phoenix-platform)
- 🐛 [Issue Tracker](https://github.com/phoenix-platform/phoenix/issues)
- 📖 [Documentation](https://phoenix-platform.io/docs)

## 🏆 Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project website

## 📜 License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.

---

Thank you for contributing to Phoenix Platform! 🚀