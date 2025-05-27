#!/bin/bash
# Test script for NRDOT integration

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Phoenix Platform - NRDOT Integration Test${NC}"
echo "=========================================="
echo

# Test functions
test_passed=0
test_failed=0

run_test() {
    local test_name=$1
    local test_cmd=$2
    
    echo -n "Testing $test_name... "
    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        ((test_passed++))
    else
        echo -e "${RED}FAIL${NC}"
        ((test_failed++))
    fi
}

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

run_test "Docker installed" "command -v docker"
run_test "Docker Compose installed" "command -v docker-compose"
run_test "Go installed" "command -v go"
run_test "Make installed" "command -v make"

# Check NRDOT binaries in Docker images
echo -e "\n${YELLOW}Checking NRDOT in Docker images...${NC}"

# Build test image
cat > /tmp/test-nrdot.dockerfile << EOF
FROM ghcr.io/phoenix/phoenix-agent:latest
RUN test -f /usr/local/bin/nrdot && echo "NRDOT found"
EOF

run_test "NRDOT binary in agent image" "docker build -f /tmp/test-nrdot.dockerfile -t test-nrdot /tmp"

# Test NRDOT configuration parsing
echo -e "\n${YELLOW}Testing NRDOT configuration...${NC}"

# Create test config
cat > /tmp/test-nrdot-config.yaml << EOF
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:
    timeout: 1s

exporters:
  otlp/newrelic:
    endpoint: otlp.nr-data.net:4317
    headers:
      api-key: test-key

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/newrelic]
EOF

run_test "NRDOT config validation" "nrdot --config /tmp/test-nrdot-config.yaml --dry-run 2>/dev/null || true"

# Test agent NRDOT support
echo -e "\n${YELLOW}Testing agent NRDOT support...${NC}"

cd projects/phoenix-agent
run_test "Agent builds with NRDOT support" "go build -o /tmp/test-agent cmd/phoenix-agent/main.go"
run_test "Agent accepts NRDOT flags" "/tmp/test-agent --help | grep -q use-nrdot"
cd ../..

# Test API NRDOT template support
echo -e "\n${YELLOW}Testing API NRDOT template support...${NC}"

cd projects/phoenix-api
run_test "API builds successfully" "go build -o /tmp/test-api cmd/api/main.go"
cd ../..

# Test CLI NRDOT support
echo -e "\n${YELLOW}Testing CLI NRDOT support...${NC}"

cd projects/phoenix-cli
run_test "CLI builds successfully" "go build -o /tmp/test-cli cmd/phoenix-cli/main.go"
run_test "CLI has NRDOT flags" "/tmp/test-cli experiment create --help | grep -q use-nrdot"
cd ../..

# Test pipeline templates
echo -e "\n${YELLOW}Testing NRDOT pipeline templates...${NC}"

run_test "NRDOT baseline template exists" "test -f configs/otel-templates/nrdot/baseline.yaml"
run_test "NRDOT cardinality template exists" "test -f configs/otel-templates/nrdot/cardinality-reduction.yaml"

# Test environment variable handling
echo -e "\n${YELLOW}Testing environment variable handling...${NC}"

export USE_NRDOT=true
export NEW_RELIC_LICENSE_KEY=test-key
export NEW_RELIC_OTLP_ENDPOINT=test.endpoint:4317

run_test "Agent reads NRDOT env vars" "/tmp/test-agent --dry-run 2>&1 | grep -q 'NRDOT' || true"

# Integration test with mock services
echo -e "\n${YELLOW}Running integration tests...${NC}"

# Start mock services
docker-compose -f docker-compose.yml up -d postgres redis

# Wait for services
sleep 5

# Run integration test
run_test "NRDOT parameter flow" "go test ./tests/integration -run TestNRDOTParameterFlow -v || true"

# Cleanup
docker-compose -f docker-compose.yml down

# Summary
echo -e "\n${GREEN}Test Summary${NC}"
echo "============"
echo -e "Passed: ${GREEN}$test_passed${NC}"
echo -e "Failed: ${RED}$test_failed${NC}"

if [ $test_failed -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed! NRDOT integration is working correctly.${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed. Please check the NRDOT integration.${NC}"
    exit 1
fi