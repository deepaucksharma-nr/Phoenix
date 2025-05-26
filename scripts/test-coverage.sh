#!/bin/bash
set -e

echo "Running tests with coverage..."

# Run tests for each project
for project in projects/*/; do
    if [ -f "$project/go.mod" ]; then
        echo "Testing $project"
        cd "$project"
        go test -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
        cd - > /dev/null
    fi
done

echo "Test coverage complete!" 