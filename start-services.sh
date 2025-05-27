#\!/bin/bash

# Kill existing processes
pkill -f phoenix-api || true
pkill -f phoenix-agent || true

# Create log directory
mkdir -p logs

# Start Phoenix API
echo "Starting Phoenix API..."
cd projects/phoenix-api
SKIP_MIGRATIONS=true ./bin/phoenix-api > ../../logs/phoenix-api.log 2>&1 &
echo "Phoenix API started (PID: $\!)"
cd ../..

# Wait for API to be ready
echo "Waiting for API to be ready..."
sleep 3

# Start Phoenix Agent
echo "Starting Phoenix Agent..."
cd projects/phoenix-agent
./bin/phoenix-agent -host-id=local-agent-1 -api-url=http://localhost:8080 > ../../logs/phoenix-agent.log 2>&1 &
echo "Phoenix Agent started (PID: $\!)"
cd ../..

# Start monitoring logs
echo -e "\n=== Monitoring logs ===\n"
echo "Phoenix API log: logs/phoenix-api.log"
echo "Phoenix Agent log: logs/phoenix-agent.log"
echo -e "\n=== Following logs (Ctrl+C to stop) ===\n"

# Tail both logs
tail -f logs/phoenix-api.log logs/phoenix-agent.log
