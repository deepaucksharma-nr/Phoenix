#!/bin/bash
# Batch Operations Example
# This demonstrates bulk operations and automation patterns

set -e

# Configuration
API_URL=${PHOENIX_API_URL:-"http://localhost:8080"}
BATCH_SIZE=5

echo "=== Phoenix Batch Operations Example ==="
echo "This example shows how to perform bulk operations efficiently"
echo

# Check authentication
if ! phoenix auth status >/dev/null 2>&1; then
    echo "Please login first:"
    phoenix auth login
fi

# 1. Create multiple experiments in batch
echo -e "\n1. Creating multiple experiments in batch..."
EXPERIMENTS=()
for i in $(seq 1 $BATCH_SIZE); do
    echo "Creating experiment batch-test-$i..."
    EXP_ID=$(phoenix experiment create \
        --name "batch-test-$i" \
        --namespace "batch-testing" \
        --pipeline-a "process-baseline-v1" \
        --pipeline-b "process-intelligent-v1" \
        --traffic-split "50/50" \
        --duration "30m" \
        --selector "app=batch-service-$i" \
        --output json | jq -r '.id')
    EXPERIMENTS+=("$EXP_ID")
done

echo "Created ${#EXPERIMENTS[@]} experiments"

# 2. Start all experiments
echo -e "\n2. Starting all experiments..."
for exp_id in "${EXPERIMENTS[@]}"; do
    phoenix experiment start "$exp_id" &
done
wait
echo "All experiments started"

# 3. Monitor all experiments in parallel
echo -e "\n3. Checking status of all experiments..."
phoenix experiment list --namespace "batch-testing" --output table

# 4. Get metrics for all running experiments
echo -e "\n4. Collecting metrics from all experiments..."
mkdir -p batch-metrics
for exp_id in "${EXPERIMENTS[@]}"; do
    echo "Fetching metrics for $exp_id..."
    phoenix experiment metrics "$exp_id" --output json > "batch-metrics/$exp_id.json"
done

# 5. Analyze metrics and make decisions
echo -e "\n5. Analyzing experiment results..."
for exp_id in "${EXPERIMENTS[@]}"; do
    METRICS_FILE="batch-metrics/$exp_id.json"
    if [ -f "$METRICS_FILE" ]; then
        # Extract key metrics
        COST_REDUCTION=$(jq -r '.summary.cost_reduction_percent // 0' "$METRICS_FILE")
        DATA_LOSS=$(jq -r '.summary.data_loss_percent // 0' "$METRICS_FILE")
        
        echo "Experiment $exp_id: Cost reduction=$COST_REDUCTION%, Data loss=$DATA_LOSS%"
        
        # Auto-promote if meets criteria
        if (( $(echo "$COST_REDUCTION > 30" | bc -l) )) && (( $(echo "$DATA_LOSS < 1" | bc -l) )); then
            echo "  -> Auto-promoting experiment $exp_id"
            phoenix experiment promote "$exp_id" --reason "Automated: Cost reduction > 30% with < 1% data loss"
        elif (( $(echo "$DATA_LOSS > 5" | bc -l) )); then
            echo "  -> Auto-stopping experiment $exp_id due to high data loss"
            phoenix experiment stop "$exp_id" --reason "Automated: Data loss exceeded 5%"
        fi
    fi
done

# 6. Bulk export configurations
echo -e "\n6. Exporting all experiment configurations..."
mkdir -p batch-configs
phoenix experiment list --namespace "batch-testing" --output json | \
    jq -r '.experiments[].id' | \
    while read -r exp_id; do
        phoenix experiment export "$exp_id" > "batch-configs/$exp_id.yaml"
    done

# 7. Generate summary report
echo -e "\n7. Generating summary report..."
cat > batch-summary.md << EOF
# Batch Operations Summary

Generated on: $(date)

## Experiments Created
Total: ${#EXPERIMENTS[@]}

## Results Summary
EOF

for exp_id in "${EXPERIMENTS[@]}"; do
    STATUS=$(phoenix experiment status "$exp_id" --output json | jq -r '.status')
    echo "- $exp_id: $STATUS" >> batch-summary.md
done

echo -e "\nSummary report saved to batch-summary.md"

# 8. Cleanup completed experiments
echo -e "\n8. Cleaning up completed experiments..."
for exp_id in "${EXPERIMENTS[@]}"; do
    STATUS=$(phoenix experiment status "$exp_id" --output json | jq -r '.status')
    if [[ "$STATUS" == "completed" ]] || [[ "$STATUS" == "failed" ]]; then
        echo "Archiving experiment $exp_id..."
        # In real scenario, you might move to archive namespace or delete
        # phoenix experiment delete "$exp_id" --force
    fi
done

echo -e "\n=== Batch Operations Complete ==="
echo "Processed ${#EXPERIMENTS[@]} experiments"
echo "Metrics saved in: ./batch-metrics/"
echo "Configurations saved in: ./batch-configs/"
echo "Summary report: ./batch-summary.md"