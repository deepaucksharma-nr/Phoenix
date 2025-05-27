#!/bin/bash
# Test script to verify pushgateway URL template rendering fix

set -e

echo "=== Testing Pushgateway URL Template Rendering Fix ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if docker-compose is running
echo "1. Checking if services are running..."
if ! docker-compose ps | grep -q "phoenix-api.*Up"; then
    echo -e "${RED}Phoenix API is not running. Please start it first.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Services are running${NC}"
echo

# Get API endpoint
API_URL="${PHOENIX_API_URL:-http://localhost:8080}"
PUSHGATEWAY_URL="${PUSHGATEWAY_URL:-http://prometheus-pushgateway:9091}"

echo "2. Creating a test deployment with baseline template..."
DEPLOYMENT_RESPONSE=$(curl -s -X POST "${API_URL}/api/v1/deployments" \
  -H "Content-Type: application/json" \
  -d '{
    "deployment_name": "test-pushgateway-deployment",
    "pipeline_name": "baseline",
    "namespace": "default",
    "target_nodes": {
      "test-node": "test-host-1"
    },
    "parameters": {
      "test_param": "test_value"
    }
  }')

DEPLOYMENT_ID=$(echo "$DEPLOYMENT_RESPONSE" | jq -r '.id')

if [ "$DEPLOYMENT_ID" == "null" ] || [ -z "$DEPLOYMENT_ID" ]; then
    echo -e "${RED}Failed to create deployment${NC}"
    echo "Response: $DEPLOYMENT_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Created deployment: $DEPLOYMENT_ID${NC}"
echo

# Wait a moment for deployment to process
sleep 2

echo "3. Fetching rendered pipeline configuration..."
CONFIG=$(curl -s "${API_URL}/api/v1/deployments/${DEPLOYMENT_ID}/config")

echo "4. Checking if METRICS_PUSHGATEWAY_URL is properly rendered..."
echo
echo "Pipeline Config (first 50 lines):"
echo "================================="
echo "$CONFIG" | head -50
echo "================================="
echo

# Check if the variable is still there (not rendered)
if echo "$CONFIG" | grep -q '${METRICS_PUSHGATEWAY_URL}'; then
    echo -e "${RED}✗ FAILED: Template variable \${METRICS_PUSHGATEWAY_URL} was not replaced${NC}"
    echo
    echo "The template still contains the unreplaced variable."
    echo "This means the fix is not working properly."
    exit 1
fi

# Check if pushgateway URL is present in the config
if echo "$CONFIG" | grep -q "$PUSHGATEWAY_URL"; then
    echo -e "${GREEN}✓ SUCCESS: Pushgateway URL was properly rendered in the config${NC}"
    echo
    echo "The config contains: $PUSHGATEWAY_URL"
else
    # Check for localhost variant
    if echo "$CONFIG" | grep -q "http://localhost:9091" || echo "$CONFIG" | grep -q "http://prometheus-pushgateway:9091"; then
        echo -e "${GREEN}✓ SUCCESS: Pushgateway URL was properly rendered in the config${NC}"
        echo
        FOUND_URL=$(echo "$CONFIG" | grep -o "http://[^/]*:9091" | head -1)
        echo "The config contains: $FOUND_URL"
    else
        echo -e "${RED}✗ WARNING: Could not find pushgateway URL in the rendered config${NC}"
        echo
        echo "Expected to find a URL like: http://prometheus-pushgateway:9091"
        echo "But the prometheusremotewrite exporter section shows:"
        echo "$CONFIG" | grep -A5 "prometheusremotewrite:" || echo "Could not find prometheusremotewrite section"
    fi
fi

echo
echo "5. Checking agent task queue..."
# Get pending tasks
TASKS=$(curl -s "${API_URL}/api/v1/tasks?status=pending&type=deployment")
TASK_COUNT=$(echo "$TASKS" | jq '. | length')

echo "Found $TASK_COUNT pending deployment tasks"

if [ "$TASK_COUNT" -gt 0 ]; then
    echo
    echo "Task details:"
    echo "$TASKS" | jq '.[0].config' | jq '{deployment_id, pushgateway_url}'
    
    # Check if pushgateway_url is in the task config
    if echo "$TASKS" | jq -r '.[0].config.pushgateway_url' | grep -q "http"; then
        echo -e "${GREEN}✓ Task includes pushgateway_url in config${NC}"
    else
        echo -e "${RED}✗ Task does not include pushgateway_url in config${NC}"
    fi
fi

echo
echo "6. Cleaning up test deployment..."
curl -s -X DELETE "${API_URL}/api/v1/deployments/${DEPLOYMENT_ID}"
echo -e "${GREEN}✓ Cleanup complete${NC}"

echo
echo "=== Test Summary ==="
echo "The fix involves:"
echo "1. Added PushgatewayURL to agent config"
echo "2. Pass pushgateway URL from API to agent in deployment tasks"
echo "3. Agent's CollectorManager now includes METRICS_PUSHGATEWAY_URL in template variables"
echo
echo "If the test passed, the template rendering is now working correctly!"