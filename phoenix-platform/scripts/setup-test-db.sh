#!/bin/bash

# Setup script for integration test database

set -e

echo "Setting up test database..."

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo "PostgreSQL is not running. Please start PostgreSQL first."
    echo ""
    echo "If you're using Docker, you can start PostgreSQL with:"
    echo "  docker run --name phoenix-test-db -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:14"
    exit 1
fi

# Create test database if it doesn't exist
PGPASSWORD=postgres psql -h localhost -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'phoenix_test'" | grep -q 1 || \
    PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE phoenix_test"

echo "Test database ready!"
echo ""
echo "To run integration tests, use:"
echo "  make test-integration"