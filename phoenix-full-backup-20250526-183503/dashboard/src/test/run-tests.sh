#!/bin/bash

# Run tests with coverage
echo "Running unit tests with coverage..."
npm test -- --coverage

# Generate coverage report
echo -e "\n\nGenerating coverage report..."
npm test -- --coverage --reporter=html

echo -e "\n\nTest Summary:"
echo "============="

# Count test files
test_count=$(find src -name "*.test.ts*" | wc -l)
echo "Total test files: $test_count"

# List tested components
echo -e "\nTested components:"
find src -name "*.test.ts*" -exec basename {} \; | sed 's/\.test\.[tj]sx\?$//' | sort | uniq

echo -e "\n\nCoverage report generated at: coverage/index.html"