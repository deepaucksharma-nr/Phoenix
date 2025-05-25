#!/bin/bash
# Phoenix CLI Example: Complete Experiment Workflow
# This script demonstrates a full experiment lifecycle using the Phoenix CLI

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Phoenix CLI Experiment Workflow Example${NC}"
echo "======================================"
echo ""

# Check if phoenix CLI is installed
if ! command -v phoenix &> /dev/null; then
    echo "Phoenix CLI not found. Please install it first."
    echo "Run: cd phoenix-platform && ./scripts/install-cli.sh"
    exit 1
fi

# Configuration
EXPERIMENT_NAME="cli-demo-$(date +%s)"
BASELINE_PIPELINE="process-baseline-v1"
CANDIDATE_PIPELINE="process-topk-v1"
TARGET_SELECTOR="app=demo,env=staging"
DURATION="10m"

echo "Experiment Configuration:"
echo "  Name:      $EXPERIMENT_NAME"
echo "  Baseline:  $BASELINE_PIPELINE"
echo "  Candidate: $CANDIDATE_PIPELINE"
echo "  Target:    $TARGET_SELECTOR"
echo "  Duration:  $DURATION"
echo ""

# Step 1: Authenticate
echo -e "${YELLOW}Step 1: Authenticating...${NC}"
# In a real scenario, you would log in interactively
# phoenix auth login

# For demo, check if already authenticated
if ! phoenix auth status &> /dev/null; then
    echo "Please authenticate first: phoenix auth login"
    exit 1
fi

echo "✓ Authentication verified"
echo ""

# Step 2: Check for overlapping experiments
echo -e "${YELLOW}Step 2: Checking for overlapping experiments...${NC}"
phoenix experiment list --status running -o json | jq -r '.[] | select(.target_nodes | to_entries | .[] | select(.key == "app" and .value == "demo")) | .name' > /tmp/overlapping.txt

if [ -s /tmp/overlapping.txt ]; then
    echo "Warning: Found overlapping experiments:"
    cat /tmp/overlapping.txt
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "✓ No overlapping experiments found"
fi
echo ""

# Step 3: Create the experiment
echo -e "${YELLOW}Step 3: Creating experiment...${NC}"
EXPERIMENT_ID=$(phoenix experiment create \
    --name "$EXPERIMENT_NAME" \
    --description "CLI workflow demonstration" \
    --baseline "$BASELINE_PIPELINE" \
    --candidate "$CANDIDATE_PIPELINE" \
    --target-selector "$TARGET_SELECTOR" \
    --duration "$DURATION" \
    --param top_k=15 \
    -o json | jq -r '.id')

if [ -z "$EXPERIMENT_ID" ]; then
    echo "Failed to create experiment"
    exit 1
fi

echo "✓ Experiment created with ID: $EXPERIMENT_ID"
echo ""

# Step 4: Start the experiment
echo -e "${YELLOW}Step 4: Starting experiment...${NC}"
phoenix experiment start "$EXPERIMENT_ID"
echo "✓ Experiment started"
echo ""

# Step 5: Monitor the experiment
echo -e "${YELLOW}Step 5: Monitoring experiment progress...${NC}"
echo "Checking status every 30 seconds for 2 minutes..."
echo ""

for i in {1..4}; do
    echo "Check $i/4:"
    phoenix experiment status "$EXPERIMENT_ID" | grep -E "Status:|Duration:|Cardinality Reduction:"
    
    # Get current metrics
    METRICS=$(phoenix experiment metrics "$EXPERIMENT_ID" -o json 2>/dev/null || echo "{}")
    if [ "$METRICS" != "{}" ]; then
        echo "  Current metrics available:"
        echo "$METRICS" | jq -r '.summary // "No summary yet"'
    fi
    
    echo ""
    
    if [ $i -lt 4 ]; then
        sleep 30
    fi
done

# Step 6: Analyze results
echo -e "${YELLOW}Step 6: Analyzing experiment results...${NC}"
phoenix experiment metrics "$EXPERIMENT_ID"
echo ""

# Step 7: Make a decision
echo -e "${YELLOW}Step 7: Decision time...${NC}"
FINAL_STATUS=$(phoenix experiment status "$EXPERIMENT_ID" -o json | jq -r '.status')

if [ "$FINAL_STATUS" == "completed" ]; then
    # Get results
    REDUCTION=$(phoenix experiment status "$EXPERIMENT_ID" -o json | jq -r '.results.cardinality_reduction // 0')
    
    echo "Experiment completed with ${REDUCTION}% cardinality reduction"
    
    if (( $(echo "$REDUCTION > 30" | bc -l) )); then
        echo "✓ Significant reduction achieved!"
        read -p "Promote candidate to production? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${YELLOW}Step 8: Promoting candidate...${NC}"
            phoenix experiment promote "$EXPERIMENT_ID" --variant candidate
            echo "✓ Candidate promoted successfully!"
        fi
    else
        echo "⚠ Reduction below threshold. Consider adjusting parameters."
    fi
else
    echo "Experiment status: $FINAL_STATUS"
    echo "For demo purposes, we'll stop here. In production, you might wait for completion."
fi

echo ""
echo -e "${GREEN}Workflow complete!${NC}"
echo ""
echo "Next steps:"
echo "  - Deploy to broader scope: phoenix pipeline deploy --name prod-$CANDIDATE_PIPELINE --pipeline $CANDIDATE_PIPELINE --selector 'env=production'"
echo "  - View all experiments: phoenix experiment list"
echo "  - Check deployment status: phoenix pipeline list-deployments"

# Cleanup
rm -f /tmp/overlapping.txt