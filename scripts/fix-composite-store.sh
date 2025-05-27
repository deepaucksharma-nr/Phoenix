#!/bin/bash
# Fix composite store type issues

cd /Users/deepaksharma/Desktop/src/Phoenix

# For CreateExperiment - remove map conversion and use TargetHosts directly
sed -i '' '35,40d' projects/phoenix-api/internal/store/composite_store.go
sed -i '' 's/TargetNodes:       targetNodes,/TargetNodes:       experiment.Config.TargetHosts,/' projects/phoenix-api/internal/store/composite_store.go

# For GetExperiment - remove the targetHosts conversion since both are []string
sed -i '' '69,73d' projects/phoenix-api/internal/store/composite_store.go
sed -i '' 's/TargetHosts: targetHosts,/TargetHosts: commonExp.TargetNodes,/' projects/phoenix-api/internal/store/composite_store.go

# For UpdateExperiment - remove map conversion
sed -i '' '/\/\/ Convert \[\]string to map\[string\]string for TargetNodes/,+5d' projects/phoenix-api/internal/store/composite_store.go
sed -i '' 's/TargetNodes:       targetNodes,/TargetNodes:       experiment.Config.TargetHosts,/' projects/phoenix-api/internal/store/composite_store.go

echo "Fixed composite store type issues"