#!/bin/bash

# Process Simulator Demo Script
# This script demonstrates how to use the Phoenix Process Simulator

echo "Phoenix Process Simulator Demo"
echo "=============================="

# Configuration
SIMULATOR_HOST="${SIMULATOR_HOST:-localhost}"
CONTROL_PORT="${CONTROL_PORT:-8090}"
METRICS_PORT="${METRICS_PORT:-8888}"

# Helper function for API calls
api_call() {
    curl -s -X $1 "http://${SIMULATOR_HOST}:${CONTROL_PORT}/api/v1/$2" \
        -H "Content-Type: application/json" \
        ${3:+-d "$3"}
}

# 1. Check simulator health
echo -e "\n1. Checking simulator health..."
api_call GET "health" | jq .

# 2. Get simulator info
echo -e "\n2. Getting simulator information..."
api_call GET "info" | jq .

# 3. Create a realistic simulation
echo -e "\n3. Creating realistic simulation (100 processes, 5 minutes)..."
SIMULATION_RESPONSE=$(api_call POST "simulations" '{
    "name": "demo-realistic",
    "type": "realistic",
    "duration": "5m",
    "parameters": {
        "process_count": 100,
        "enable_chaos": false
    }
}')
echo "$SIMULATION_RESPONSE" | jq .

# Extract simulation ID
SIM_ID=$(echo "$SIMULATION_RESPONSE" | jq -r '.ID')
echo "Simulation ID: $SIM_ID"

# 4. Start the simulation
echo -e "\n4. Starting simulation..."
api_call POST "simulations/$SIM_ID/start" | jq .

# 5. Wait and check metrics
echo -e "\n5. Waiting 30 seconds for processes to stabilize..."
sleep 30

echo -e "\n6. Checking Prometheus metrics..."
echo "Process count:"
curl -s "http://${SIMULATOR_HOST}:${METRICS_PORT}/metrics" | grep "phoenix_simulator_process_count"

echo -e "\nSample process metrics:"
curl -s "http://${SIMULATOR_HOST}:${METRICS_PORT}/metrics" | grep "process_cpu_seconds_total" | head -5

# 7. Trigger a chaos event (CPU spike)
echo -e "\n7. Triggering CPU spike in Python apps..."
api_call POST "chaos/cpu-spike" '{
    "process_pattern": "python-app",
    "duration": "30s",
    "intensity": 85.0
}' | jq .

# 8. Get simulation status
echo -e "\n8. Getting simulation status..."
api_call GET "simulations/$SIM_ID" | jq .

# 9. Demo high-cardinality simulation
echo -e "\n9. Creating high-cardinality simulation (500 processes)..."
HC_RESPONSE=$(api_call POST "simulations" '{
    "name": "demo-high-cardinality",
    "type": "high-cardinality",
    "duration": "2m",
    "parameters": {
        "process_count": 500,
        "enable_chaos": false
    }
}')
HC_SIM_ID=$(echo "$HC_RESPONSE" | jq -r '.ID')

api_call POST "simulations/$HC_SIM_ID/start" | jq .

# 10. Compare cardinality
echo -e "\n10. Waiting 30 seconds then comparing cardinality..."
sleep 30

echo "Unique time series count:"
SERIES_COUNT=$(curl -s "http://${SIMULATOR_HOST}:${METRICS_PORT}/metrics" | grep -E "^process_" | wc -l)
echo "Total unique metric series: $SERIES_COUNT"

# 11. Stop simulations
echo -e "\n11. Stopping simulations..."
api_call POST "simulations/$SIM_ID/stop" | jq .
api_call POST "simulations/$HC_SIM_ID/stop" | jq .

# 12. Get final results
echo -e "\n12. Getting simulation results..."
echo "Realistic simulation results:"
api_call GET "simulations/$SIM_ID/results" | jq .

echo -e "\nHigh-cardinality simulation results:"
api_call GET "simulations/$HC_SIM_ID/results" | jq .

echo -e "\n✅ Demo completed!"
echo "=============================="
echo "Key observations:"
echo "- Process simulator can create various workload patterns"
echo "- Metrics are exposed in Prometheus format"
echo "- Chaos engineering can be triggered on demand"
echo "- High-cardinality scenarios create many unique time series"
echo ""
echo "Next steps:"
echo "1. Run a Phoenix experiment against this simulator"
echo "2. Compare baseline vs optimized pipeline metrics"
echo "3. Validate cardinality reduction and process retention"