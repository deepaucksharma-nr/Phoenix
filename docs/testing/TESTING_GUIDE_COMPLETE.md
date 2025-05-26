# Phoenix Platform Testing Guide - Complete Documentation

## 📋 Table of Contents

1. [Testing Overview](#testing-overview)
2. [Testing Strategy](#testing-strategy)
3. [Unit Testing](#unit-testing)
4. [Integration Testing](#integration-testing)
5. [End-to-End Testing](#end-to-end-testing)
6. [Performance Testing](#performance-testing)
7. [Security Testing](#security-testing)
8. [Test Automation](#test-automation)
9. [Test Data Management](#test-data-management)
10. [Testing Best Practices](#testing-best-practices)

---

## Testing Overview

The Phoenix Platform implements a comprehensive testing strategy ensuring quality, reliability, and performance across all components. Our testing philosophy emphasizes:

- **Test-Driven Development (TDD)** for new features
- **Comprehensive Coverage** at all levels
- **Automated Testing** in CI/CD pipeline
- **Fast Feedback** loops for developers
- **Production-Like** test environments

### Testing Pyramid
```
         ╱─────────╲
        ╱   E2E     ╲      (10%)
       ╱   Tests     ╲
      ╱───────────────╲
     ╱  Integration    ╲    (20%)
    ╱     Tests         ╲
   ╱─────────────────────╲
  ╱     Unit Tests        ╲  (70%)
 ╱─────────────────────────╲
```

---

## Testing Strategy

### Test Levels and Scope

| Level | Scope | Speed | Confidence | Maintenance |
|-------|-------|-------|------------|-------------|
| Unit | Single function/method | Fast (ms) | Low | Low |
| Integration | Multiple components | Medium (s) | Medium | Medium |
| E2E | Full user journey | Slow (min) | High | High |
| Performance | System limits | Slow | Specific | Medium |

### Coverage Requirements

- **Unit Tests**: Minimum 80% coverage
- **Integration Tests**: Critical paths covered
- **E2E Tests**: Key user journeys
- **Performance Tests**: Load and stress scenarios

---

## Unit Testing

### Go Unit Testing

#### Basic Test Structure
```go
package service_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/phoenix/platform/projects/api/internal/service"
)

func TestExperimentService_Create(t *testing.T) {
    // Arrange
    svc := service.NewExperimentService()
    input := service.CreateExperimentInput{
        Name: "Test Experiment",
        Type: "AB_TEST",
    }

    // Act
    result, err := svc.Create(context.Background(), input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "Test Experiment", result.Name)
    assert.NotEmpty(t, result.ID)
}
```

#### Table-Driven Tests
```go
func TestValidateExperimentName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid name",
            input:   "My Experiment",
            wantErr: false,
        },
        {
            name:    "empty name",
            input:   "",
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name:    "name too long",
            input:   strings.Repeat("a", 101),
            wantErr: true,
            errMsg:  "name must be less than 100 characters",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateExperimentName(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

#### Mocking Dependencies
```go
//go:generate mockgen -source=interfaces.go -destination=mocks/mock_interfaces.go

func TestController_ProcessExperiment(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Create mocks
    mockStore := mocks.NewMockExperimentStore(ctrl)
    mockClient := mocks.NewMockGeneratorClient(ctrl)

    // Set expectations
    mockStore.EXPECT().
        GetByID(gomock.Any(), "exp-123").
        Return(&Experiment{ID: "exp-123", Status: "PENDING"}, nil)

    mockClient.EXPECT().
        GeneratePipeline(gomock.Any(), gomock.Any()).
        Return(&Pipeline{ID: "pipe-456"}, nil)

    // Test
    controller := NewController(mockStore, mockClient)
    err := controller.ProcessExperiment(context.Background(), "exp-123")
    
    assert.NoError(t, err)
}
```

### TypeScript/React Unit Testing

#### Component Testing
```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { ExperimentCard } from './ExperimentCard';

describe('ExperimentCard', () => {
  const mockExperiment = {
    id: '123',
    name: 'Test Experiment',
    status: 'RUNNING',
    costReduction: 45.2,
  };

  it('renders experiment information', () => {
    render(<ExperimentCard experiment={mockExperiment} />);
    
    expect(screen.getByText('Test Experiment')).toBeInTheDocument();
    expect(screen.getByText('RUNNING')).toBeInTheDocument();
    expect(screen.getByText('45.2%')).toBeInTheDocument();
  });

  it('calls onSelect when clicked', () => {
    const handleSelect = jest.fn();
    render(
      <ExperimentCard 
        experiment={mockExperiment} 
        onSelect={handleSelect}
      />
    );
    
    fireEvent.click(screen.getByRole('button'));
    expect(handleSelect).toHaveBeenCalledWith('123');
  });
});
```

#### Hook Testing
```typescript
import { renderHook, act } from '@testing-library/react-hooks';
import { useExperimentData } from './useExperimentData';

describe('useExperimentData', () => {
  it('fetches experiment data on mount', async () => {
    const { result, waitForNextUpdate } = renderHook(() => 
      useExperimentData('123')
    );

    expect(result.current.loading).toBe(true);
    
    await waitForNextUpdate();
    
    expect(result.current.loading).toBe(false);
    expect(result.current.data).toEqual({
      id: '123',
      name: 'Test Experiment',
    });
  });

  it('handles errors gracefully', async () => {
    // Mock API to return error
    jest.spyOn(api, 'getExperiment').mockRejectedValue(
      new Error('Network error')
    );

    const { result, waitForNextUpdate } = renderHook(() => 
      useExperimentData('invalid')
    );

    await waitForNextUpdate();
    
    expect(result.current.error).toBe('Network error');
    expect(result.current.data).toBeNull();
  });
});
```

---

## Integration Testing

### Service Integration Tests
```go
// tests/integration/experiment_flow_test.go
package integration

import (
    "testing"
    "github.com/phoenix/platform/tests/integration/helpers"
)

func TestExperimentLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test environment
    env := helpers.NewTestEnvironment(t)
    defer env.Cleanup()

    // Create API client
    client := env.APIClient()

    // Test experiment creation
    t.Run("CreateExperiment", func(t *testing.T) {
        exp, err := client.CreateExperiment(&CreateExperimentRequest{
            Name: "Integration Test",
            Type: "AB_TEST",
        })
        require.NoError(t, err)
        assert.NotEmpty(t, exp.ID)
    })

    // Test experiment start
    t.Run("StartExperiment", func(t *testing.T) {
        err := client.StartExperiment(exp.ID)
        require.NoError(t, err)

        // Wait for deployment
        require.Eventually(t, func() bool {
            status, _ := client.GetExperimentStatus(exp.ID)
            return status == "RUNNING"
        }, 30*time.Second, 1*time.Second)
    })

    // Test metrics collection
    t.Run("CollectMetrics", func(t *testing.T) {
        // Simulate traffic
        env.GenerateTraffic(exp.ID, 100)

        // Check metrics
        metrics, err := client.GetMetrics(exp.ID)
        require.NoError(t, err)
        assert.Greater(t, metrics.RequestCount, int64(0))
    })
}
```

### Database Integration Tests
```go
func TestPostgresStore_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database integration test")
    }

    // Setup test database
    db := testdb.Setup(t)
    defer testdb.Cleanup(t, db)

    store := postgres.NewStore(db)

    t.Run("CreateAndRetrieve", func(t *testing.T) {
        // Create
        exp := &Experiment{
            Name: "Test Experiment",
            Type: "CANARY",
        }
        err := store.Create(context.Background(), exp)
        require.NoError(t, err)

        // Retrieve
        retrieved, err := store.GetByID(context.Background(), exp.ID)
        require.NoError(t, err)
        assert.Equal(t, exp.Name, retrieved.Name)
    })

    t.Run("ConcurrentOperations", func(t *testing.T) {
        var wg sync.WaitGroup
        errors := make(chan error, 10)

        // Concurrent creates
        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func(i int) {
                defer wg.Done()
                err := store.Create(context.Background(), &Experiment{
                    Name: fmt.Sprintf("Concurrent %d", i),
                })
                if err != nil {
                    errors <- err
                }
            }(i)
        }

        wg.Wait()
        close(errors)

        // Check for errors
        for err := range errors {
            t.Errorf("Concurrent operation failed: %v", err)
        }
    })
}
```

---

## End-to-End Testing

### E2E Test Framework
```typescript
// tests/e2e/experiment.spec.ts
import { test, expect } from '@playwright/test';
import { createExperiment, waitForStatus } from './helpers';

test.describe('Experiment Management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.fill('[data-testid="email"]', 'test@phoenix.io');
    await page.fill('[data-testid="password"]', 'password');
    await page.click('[data-testid="login-button"]');
  });

  test('complete experiment workflow', async ({ page }) => {
    // Create experiment
    await page.click('[data-testid="create-experiment"]');
    await page.fill('[data-testid="experiment-name"]', 'E2E Test');
    await page.selectOption('[data-testid="experiment-type"]', 'AB_TEST');
    await page.click('[data-testid="submit"]');

    // Verify creation
    await expect(page.locator('.success-toast')).toContainText(
      'Experiment created successfully'
    );

    // Start experiment
    await page.click('[data-testid="experiment-row-E2E Test"]');
    await page.click('[data-testid="start-experiment"]');
    
    // Wait for running status
    await waitForStatus(page, 'E2E Test', 'RUNNING');

    // Check metrics
    await page.click('[data-testid="metrics-tab"]');
    await page.waitForSelector('[data-testid="metrics-chart"]');
    
    // Verify data appears
    await expect(page.locator('[data-testid="cost-reduction"]'))
      .toContainText(/\d+\.\d+%/);

    // Stop experiment
    await page.click('[data-testid="stop-experiment"]');
    await page.fill('[data-testid="stop-reason"]', 'E2E test complete');
    await page.click('[data-testid="confirm-stop"]');

    // Verify stopped
    await waitForStatus(page, 'E2E Test', 'COMPLETED');
  });

  test('error handling', async ({ page }) => {
    // Try to create invalid experiment
    await page.click('[data-testid="create-experiment"]');
    await page.click('[data-testid="submit"]'); // No data

    // Verify validation errors
    await expect(page.locator('.field-error')).toContainText(
      'Name is required'
    );
  });
});
```

### API E2E Tests
```bash
#!/bin/bash
# tests/e2e/api-e2e.sh

# Setup
API_URL=${API_URL:-http://localhost:8080}
AUTH_TOKEN=$(curl -s -X POST $API_URL/auth/login \
  -d '{"email":"test@phoenix.io","password":"password"}' | jq -r .token)

# Test 1: Create Experiment
echo "Test 1: Creating experiment..."
EXPERIMENT_ID=$(curl -s -X POST $API_URL/api/v1/experiments \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API E2E Test",
    "type": "AB_TEST",
    "config": {
      "duration": "1h",
      "traffic_split": {"baseline": 50, "candidate": 50}
    }
  }' | jq -r .id)

# Test 2: Start Experiment
echo "Test 2: Starting experiment..."
curl -X POST $API_URL/api/v1/experiments/$EXPERIMENT_ID/start \
  -H "Authorization: Bearer $AUTH_TOKEN"

# Test 3: Check Status
echo "Test 3: Checking status..."
STATUS=$(curl -s $API_URL/api/v1/experiments/$EXPERIMENT_ID \
  -H "Authorization: Bearer $AUTH_TOKEN" | jq -r .status)

if [ "$STATUS" != "RUNNING" ]; then
  echo "Error: Expected status RUNNING, got $STATUS"
  exit 1
fi

echo "All E2E tests passed!"
```

---

## Performance Testing

### Load Testing with K6
```javascript
// tests/performance/load/experiment-api.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp to 200
    { duration: '5m', target: 200 },  // Stay at 200
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.1'],    // Error rate under 10%
  },
};

export default function() {
  // Login
  let loginRes = http.post('http://localhost:8080/auth/login', {
    email: 'test@phoenix.io',
    password: 'password',
  });
  
  check(loginRes, {
    'login successful': (r) => r.status === 200,
  });
  
  let authToken = loginRes.json('token');
  
  // Create experiment
  let payload = JSON.stringify({
    name: `Load Test ${Date.now()}`,
    type: 'AB_TEST',
  });
  
  let params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authToken}`,
    },
  };
  
  let res = http.post(
    'http://localhost:8080/api/v1/experiments',
    payload,
    params
  );
  
  check(res, {
    'experiment created': (r) => r.status === 201,
    'has id': (r) => r.json('id') !== '',
  });
  
  sleep(1);
}
```

### Stress Testing
```go
// tests/performance/stress/memory_stress_test.go
func TestMemoryUnderStress(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test")
    }

    // Monitor memory
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    initialMem := m.Alloc

    // Create stress
    experiments := make([]*Experiment, 10000)
    for i := 0; i < 10000; i++ {
        experiments[i] = &Experiment{
            ID:   fmt.Sprintf("exp-%d", i),
            Name: fmt.Sprintf("Stress Test %d", i),
            Data: make([]byte, 1024), // 1KB per experiment
        }
    }

    // Check memory growth
    runtime.ReadMemStats(&m)
    memGrowth := m.Alloc - initialMem
    memGrowthMB := float64(memGrowth) / 1024 / 1024

    assert.Less(t, memGrowthMB, 100.0, "Memory growth exceeded 100MB")

    // Cleanup and verify GC
    experiments = nil
    runtime.GC()
    runtime.ReadMemStats(&m)
    
    finalMem := m.Alloc
    assert.Less(t, finalMem, initialMem*2, "Memory not properly released")
}
```

---

## Security Testing

### Security Test Suite
```go
// tests/security/injection_test.go
func TestSQLInjectionPrevention(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {
            name:  "basic injection",
            input: "'; DROP TABLE experiments; --",
        },
        {
            name:  "union injection",
            input: "' UNION SELECT * FROM users --",
        },
        {
            name:  "blind injection",
            input: "' OR '1'='1",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Attempt injection
            _, err := store.Search(tt.input)
            
            // Should not return SQL error
            if err != nil {
                assert.NotContains(t, err.Error(), "SQL")
                assert.NotContains(t, err.Error(), "syntax")
            }
            
            // Verify table still exists
            var count int
            err = db.QueryRow("SELECT COUNT(*) FROM experiments").Scan(&count)
            assert.NoError(t, err)
        })
    }
}
```

### Authentication Testing
```typescript
// tests/security/auth.test.ts
describe('Authentication Security', () => {
  test('prevents access without token', async () => {
    const response = await fetch('/api/v1/experiments');
    expect(response.status).toBe(401);
  });

  test('rejects invalid tokens', async () => {
    const response = await fetch('/api/v1/experiments', {
      headers: {
        'Authorization': 'Bearer invalid-token',
      },
    });
    expect(response.status).toBe(401);
  });

  test('enforces token expiration', async () => {
    // Use expired token
    const expiredToken = generateExpiredToken();
    const response = await fetch('/api/v1/experiments', {
      headers: {
        'Authorization': `Bearer ${expiredToken}`,
      },
    });
    expect(response.status).toBe(401);
  });
});
```

---

## Test Automation

### CI/CD Integration
```yaml
# .github/workflows/test.yml
name: Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [api, controller, generator, analytics]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: |
          cd projects/${{ matrix.service }}
          go test -v -race -coverprofile=coverage.out ./...
          
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./projects/${{ matrix.service }}/coverage.out
          flags: ${{ matrix.service }}

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      - name: Run integration tests
        run: |
          make test-integration

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Start services
        run: docker-compose up -d
        
      - name: Wait for services
        run: ./scripts/wait-for-services.sh
        
      - name: Run E2E tests
        run: |
          cd tests/e2e
          npm install
          npx playwright test
```

### Test Reporting
```xml
<!-- Example test report format -->
<testsuites>
  <testsuite name="ExperimentService" tests="15" failures="0" time="2.341">
    <testcase name="TestCreate" time="0.102"/>
    <testcase name="TestUpdate" time="0.098"/>
    <testcase name="TestDelete" time="0.087"/>
  </testsuite>
</testsuites>
```

---

## Test Data Management

### Test Data Generation
```go
// tests/fixtures/generator.go
package fixtures

func GenerateExperiment(opts ...Option) *Experiment {
    cfg := &config{
        name: faker.Name(),
        type: "AB_TEST",
    }
    
    for _, opt := range opts {
        opt(cfg)
    }
    
    return &Experiment{
        ID:        uuid.New().String(),
        Name:      cfg.name,
        Type:      cfg.type,
        CreatedAt: time.Now(),
    }
}

// Usage
exp := fixtures.GenerateExperiment(
    fixtures.WithName("Test Experiment"),
    fixtures.WithType("CANARY"),
)
```

### Test Database Management
```bash
#!/bin/bash
# scripts/test-db.sh

# Create test database
createdb phoenix_test

# Run migrations
migrate -database postgres://localhost/phoenix_test up

# Seed test data
psql phoenix_test < tests/fixtures/seed.sql

# Cleanup after tests
dropdb phoenix_test
```

---

## Testing Best Practices

### General Guidelines

1. **Test Naming**
   ```go
   // Good
   func TestExperimentService_Create_WithValidInput_ReturnsExperiment(t *testing.T)
   
   // Bad
   func TestCreate(t *testing.T)
   ```

2. **Test Organization**
   ```
   project/
   ├── internal/
   │   ├── service/
   │   │   ├── experiment.go
   │   │   └── experiment_test.go
   │   └── store/
   │       ├── postgres.go
   │       └── postgres_test.go
   └── tests/
       ├── integration/
       └── fixtures/
   ```

3. **Test Independence**
   - Each test should be independent
   - Use fresh test data
   - Clean up after tests
   - Avoid shared state

4. **Test Speed**
   - Use `t.Parallel()` for independent tests
   - Mock external dependencies
   - Use in-memory databases for unit tests
   - Skip slow tests with `-short` flag

5. **Test Clarity**
   - Clear test names
   - Arrange-Act-Assert pattern
   - Descriptive assertions
   - Helpful error messages

### Common Patterns

#### Test Helpers
```go
func setupTest(t *testing.T) (*Service, func()) {
    t.Helper()
    
    // Setup
    db := setupTestDB(t)
    svc := NewService(db)
    
    // Cleanup function
    cleanup := func() {
        db.Close()
    }
    
    return svc, cleanup
}

// Usage
func TestSomething(t *testing.T) {
    svc, cleanup := setupTest(t)
    defer cleanup()
    
    // Test code
}
```

#### Custom Assertions
```go
func assertExperimentEqual(t *testing.T, expected, actual *Experiment) {
    t.Helper()
    
    assert.Equal(t, expected.Name, actual.Name)
    assert.Equal(t, expected.Type, actual.Type)
    assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
}
```

---

## Test Metrics and Reports

### Coverage Goals
- Overall: > 80%
- Critical paths: > 90%
- New code: > 85%

### Test Execution Time
- Unit tests: < 2 minutes
- Integration tests: < 10 minutes
- E2E tests: < 30 minutes
- Full suite: < 45 minutes

### Quality Metrics
- Flaky test rate: < 1%
- Test maintenance time: < 10% of dev time
- Bug escape rate: < 5%

---

*This testing guide provides comprehensive strategies for ensuring quality in the Phoenix Platform.*  
*Keep tests fast, reliable, and maintainable for long-term success.*  
*Last Updated: May 2025*