#!/bin/bash
# Start demo Phoenix agents for UI testing

set -e

NUM_AGENTS=${1:-5}
API_URL=${PHOENIX_API_URL:-http://localhost:8080}

echo "Starting $NUM_AGENTS demo agents..."

for i in $(seq 1 $NUM_AGENTS); do
  HOST_ID="demo-host-$(printf "%03d" $i)"
  GROUP="demo-group-$((($i - 1) / 3 + 1))"
  
  # Start agent in background
  PHOENIX_AGENT_ID=$HOST_ID \
  PHOENIX_AGENT_GROUP=$GROUP \
  PHOENIX_API_URL=$API_URL \
  ./projects/phoenix-agent/phoenix-agent \
    --log-level=info \
    --poll-interval=30s \
    --host-tags="env=demo,group=$GROUP,zone=us-east-$((($i - 1) % 3 + 1))" &
    
  echo "✓ Started agent: $HOST_ID (group: $GROUP)"
done

echo "All demo agents started successfully!"