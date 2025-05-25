#!/bin/bash
# CI/CD Integration Example
# This demonstrates how to integrate Phoenix CLI into CI/CD pipelines

set -e

# CI/CD environment variables
CI_COMMIT_SHA=${CI_COMMIT_SHA:-$(git rev-parse HEAD)}
CI_BRANCH=${CI_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}
CI_BUILD_ID=${CI_BUILD_ID:-"local-$(date +%s)"}
ENVIRONMENT=${ENVIRONMENT:-"staging"}

echo "=== Phoenix CI/CD Integration Example ==="
echo "Build: $CI_BUILD_ID"
echo "Commit: $CI_COMMIT_SHA"
echo "Branch: $CI_BRANCH"
echo "Environment: $ENVIRONMENT"
echo

# Function to send notifications (Slack, email, etc.)
notify() {
    local level=$1
    local message=$2
    echo "[$level] $message"
    # In real CI/CD, integrate with notification service
    # curl -X POST $SLACK_WEBHOOK -d "{\"text\":\"[$level] $message\"}"
}

# 1. Setup authentication using CI/CD secrets
echo "1. Authenticating with Phoenix API..."
if [ -n "$PHOENIX_API_TOKEN" ]; then
    # Use token directly from CI/CD secrets
    export PHOENIX_AUTH_TOKEN="$PHOENIX_API_TOKEN"
    echo "Using API token from environment"
else
    # Fallback to interactive login (for local testing)
    phoenix auth login
fi

# 2. Validate pipeline configurations
echo -e "\n2. Validating pipeline configurations..."
PIPELINE_CONFIGS=(
    "pipelines/production/baseline.yaml"
    "pipelines/production/optimized.yaml"
)

for config in "${PIPELINE_CONFIGS[@]}"; do
    if [ -f "$config" ]; then
        echo "Validating $config..."
        phoenix pipeline validate --file "$config"
    fi
done

# 3. Create experiment based on branch/environment
echo -e "\n3. Creating CI/CD experiment..."
EXPERIMENT_NAME="ci-${CI_BUILD_ID}-${CI_BRANCH//\//-}"
EXPERIMENT_NAME=${EXPERIMENT_NAME:0:63} # Kubernetes name limit

# Different strategies for different branches
if [[ "$CI_BRANCH" == "main" ]] || [[ "$CI_BRANCH" == "master" ]]; then
    TRAFFIC_SPLIT="90/10"  # Conservative for production
    DURATION="2h"
    NAMESPACE="production"
elif [[ "$CI_BRANCH" == "staging" ]]; then
    TRAFFIC_SPLIT="50/50"  # Balanced for staging
    DURATION="1h"
    NAMESPACE="staging"
else
    TRAFFIC_SPLIT="20/80"  # Aggressive for feature branches
    DURATION="30m"
    NAMESPACE="development"
fi

EXPERIMENT_ID=$(phoenix experiment create \
    --name "$EXPERIMENT_NAME" \
    --namespace "$NAMESPACE" \
    --pipeline-a "process-baseline-v1" \
    --pipeline-b "process-optimized-${CI_BRANCH}" \
    --traffic-split "$TRAFFIC_SPLIT" \
    --duration "$DURATION" \
    --selector "app=${CI_BRANCH}-service" \
    --metadata "{\"build_id\":\"$CI_BUILD_ID\",\"commit\":\"$CI_COMMIT_SHA\",\"branch\":\"$CI_BRANCH\"}" \
    --output json | jq -r '.id')

notify "INFO" "Created experiment $EXPERIMENT_ID for build $CI_BUILD_ID"

# 4. Start experiment and wait for initial metrics
echo -e "\n4. Starting experiment..."
phoenix experiment start "$EXPERIMENT_ID"

# 5. Run automated tests during experiment
echo -e "\n5. Running integration tests..."
TEST_RESULTS_FILE="test-results-$CI_BUILD_ID.json"

# Simulate running tests (replace with actual test command)
cat > "$TEST_RESULTS_FILE" << EOF
{
    "passed": 45,
    "failed": 0,
    "duration": "2m30s",
    "metrics": {
        "p99_latency": 250,
        "error_rate": 0.001,
        "throughput": 5000
    }
}
EOF

TEST_PASSED=$(jq -r '.failed == 0' "$TEST_RESULTS_FILE")

# 6. Monitor experiment with quality gates
echo -e "\n6. Monitoring experiment with quality gates..."
MONITORING_DURATION=300  # 5 minutes
START_TIME=$(date +%s)
QUALITY_GATE_PASSED=true

while [ $(($(date +%s) - START_TIME)) -lt $MONITORING_DURATION ]; do
    # Get current metrics
    METRICS=$(phoenix experiment metrics "$EXPERIMENT_ID" --output json)
    
    # Check quality gates
    COST_REDUCTION=$(echo "$METRICS" | jq -r '.summary.cost_reduction_percent // 0')
    DATA_LOSS=$(echo "$METRICS" | jq -r '.summary.data_loss_percent // 0')
    ERROR_RATE=$(echo "$METRICS" | jq -r '.pipeline_b.error_rate // 0')
    
    echo "Current metrics: Cost reduction=$COST_REDUCTION%, Data loss=$DATA_LOSS%, Error rate=$ERROR_RATE"
    
    # Quality gate checks
    if (( $(echo "$DATA_LOSS > 2" | bc -l) )); then
        QUALITY_GATE_PASSED=false
        notify "ERROR" "Quality gate failed: Data loss $DATA_LOSS% exceeds 2% threshold"
        break
    fi
    
    if (( $(echo "$ERROR_RATE > 0.01" | bc -l) )); then
        QUALITY_GATE_PASSED=false
        notify "ERROR" "Quality gate failed: Error rate $ERROR_RATE exceeds 1% threshold"
        break
    fi
    
    sleep 30
done

# 7. Make deployment decision
echo -e "\n7. Making deployment decision..."
if [[ "$TEST_PASSED" == "true" ]] && [[ "$QUALITY_GATE_PASSED" == "true" ]]; then
    # Check if cost reduction meets threshold
    FINAL_METRICS=$(phoenix experiment metrics "$EXPERIMENT_ID" --output json)
    COST_REDUCTION=$(echo "$FINAL_METRICS" | jq -r '.summary.cost_reduction_percent // 0')
    
    if (( $(echo "$COST_REDUCTION > 20" | bc -l) )); then
        echo "Promoting experiment - all criteria met"
        phoenix experiment promote "$EXPERIMENT_ID" \
            --reason "CI/CD: Build $CI_BUILD_ID passed all quality gates with $COST_REDUCTION% cost reduction"
        
        notify "SUCCESS" "Experiment $EXPERIMENT_ID promoted successfully"
        
        # Tag the commit
        git tag -a "phoenix-promoted-$CI_BUILD_ID" -m "Phoenix experiment $EXPERIMENT_ID promoted"
        
        EXIT_CODE=0
    else
        echo "Stopping experiment - insufficient cost reduction"
        phoenix experiment stop "$EXPERIMENT_ID" \
            --reason "CI/CD: Cost reduction $COST_REDUCTION% below 20% threshold"
        
        notify "INFO" "Experiment $EXPERIMENT_ID stopped - insufficient cost reduction"
        EXIT_CODE=0
    fi
else
    echo "Stopping experiment - quality gates failed"
    phoenix experiment stop "$EXPERIMENT_ID" \
        --reason "CI/CD: Quality gates failed - Tests: $TEST_PASSED, Gates: $QUALITY_GATE_PASSED"
    
    notify "ERROR" "Experiment $EXPERIMENT_ID failed quality gates"
    EXIT_CODE=1
fi

# 8. Generate and upload reports
echo -e "\n8. Generating CI/CD reports..."
REPORT_DIR="phoenix-reports-$CI_BUILD_ID"
mkdir -p "$REPORT_DIR"

# Export experiment details
phoenix experiment export "$EXPERIMENT_ID" > "$REPORT_DIR/experiment.yaml"
phoenix experiment metrics "$EXPERIMENT_ID" --output json > "$REPORT_DIR/metrics.json"

# Generate summary report
cat > "$REPORT_DIR/summary.md" << EOF
# Phoenix CI/CD Report

**Build ID:** $CI_BUILD_ID  
**Commit:** $CI_COMMIT_SHA  
**Branch:** $CI_BRANCH  
**Experiment:** $EXPERIMENT_ID  
**Result:** $([ $EXIT_CODE -eq 0 ] && echo "SUCCESS" || echo "FAILED")

## Quality Gates
- Tests Passed: $TEST_PASSED
- Quality Gates: $QUALITY_GATE_PASSED
- Cost Reduction: $COST_REDUCTION%
- Data Loss: $DATA_LOSS%

## Decision
$([ $EXIT_CODE -eq 0 ] && echo "Experiment promoted to production" || echo "Experiment failed quality checks")

Generated at: $(date)
EOF

# Upload artifacts (platform-specific)
echo "Reports generated in $REPORT_DIR/"
# In real CI/CD:
# - GitLab: Use artifacts directive
# - Jenkins: Use archiveArtifacts
# - GitHub Actions: Use upload-artifact action

echo -e "\n=== CI/CD Integration Complete ==="
exit $EXIT_CODE