#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cleanup() {
    echo "Cleaning up test environment..."
    docker-compose -f "$PROJECT_ROOT/test/docker-compose.test.yml" down -v || true
    
    # Kill any background processes
    if [ ! -z "$API_PID" ]; then
        kill $API_PID 2>/dev/null || true
    fi
    if [ ! -z "$GENERATOR_PID" ]; then
        kill $GENERATOR_PID 2>/dev/null || true
    fi
}

trap cleanup EXIT

echo "Phoenix Platform Integration Tests"
echo "================================="

# Start test dependencies (PostgreSQL)
echo "Starting test database..."
cat > "$PROJECT_ROOT/test/docker-compose.test.yml" << EOF
version: '3.8'
services:
  postgres-test:
    image: postgres:15
    environment:
      POSTGRES_DB: phoenix_test
      POSTGRES_USER: phoenix
      POSTGRES_PASSWORD: testpass
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data
    command: postgres -c fsync=off -c synchronous_commit=off -c full_page_writes=off
EOF

docker-compose -f "$PROJECT_ROOT/test/docker-compose.test.yml" up -d postgres-test

# Wait for database to be ready
echo "Waiting for database to be ready..."
for i in {1..30}; do
    if docker-compose -f "$PROJECT_ROOT/test/docker-compose.test.yml" exec -T postgres-test pg_isready -U phoenix -d phoenix_test > /dev/null 2>&1; then
        echo "Database is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Database failed to start within 30 seconds"
        exit 1
    fi
    sleep 1
done

# Set environment variables for tests
export DATABASE_URL="postgres://phoenix:testpass@localhost:5433/phoenix_test?sslmode=disable"
export JWT_SECRET="test-jwt-secret-for-integration-tests"
export GIT_TOKEN="test-git-token"
export NEW_RELIC_API_KEY="test-api-key"

# Build test binaries
echo "Building test services..."
cd "$PROJECT_ROOT"

echo "Building API service..."
go build -o bin/api-test cmd/api/main.go

echo "Building Generator service..."
go build -o bin/generator-test cmd/generator/main.go

# Start services in background
echo "Starting test services..."

echo "Starting API service on port 8081..."
HTTP_PORT=8081 GRPC_PORT=5051 ./bin/api-test &
API_PID=$!

echo "Starting Generator service on port 8083..."
HTTP_PORT=:8083 ./bin/generator-test &
GENERATOR_PID=$!

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 5

# Check if services are running
if ! kill -0 $API_PID 2>/dev/null; then
    echo "API service failed to start"
    exit 1
fi

if ! kill -0 $GENERATOR_PID 2>/dev/null; then
    echo "Generator service failed to start"
    exit 1
fi

# Wait for HTTP endpoints to be ready
echo "Waiting for API endpoints to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8081/health > /dev/null 2>&1; then
        echo "API service is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "API service health check failed"
        exit 1
    fi
    sleep 1
done

echo "Waiting for Generator endpoints to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:8083/health > /dev/null 2>&1; then
        echo "Generator service is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Generator service health check failed"
        exit 1
    fi
    sleep 1
done

# Run integration tests
echo "Running E2E tests..."
export API_URL="http://localhost:8081"
export GENERATOR_URL="http://localhost:8083"

cd "$PROJECT_ROOT"

# Test categories
test_categories=(
    "SimpleE2E"
)

failed_tests=()
passed_tests=()

for category in "${test_categories[@]}"; do
    echo -e "\n--- Testing: $category ---"
    if go test -v -tags=e2e ./test/e2e/... -run "$category" -timeout=10m; then
        passed_tests+=($category)
        echo "✓ $category tests passed"
    else
        failed_tests+=($category)
        echo "✗ $category tests failed"
    fi
done

# Summary
echo -e "\n\nTest Summary"
echo "============"
echo "Passed: ${#passed_tests[@]} (${passed_tests[*]})"
echo "Failed: ${#failed_tests[@]} (${failed_tests[*]})"

if [ ${#failed_tests[@]} -gt 0 ]; then
    echo -e "\nSome tests failed!"
    exit 1
else
    echo -e "\nAll integration tests passed!"
fi

echo "Integration tests completed successfully!"