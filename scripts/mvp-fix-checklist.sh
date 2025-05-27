#!/bin/bash
# Phoenix MVP Fix Checklist
# Interactive script to guide through fixing identified issues

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔧 Phoenix MVP Fix Checklist"
echo "============================"
echo ""

# Function to prompt for confirmation
confirm() {
    local prompt=$1
    read -p "$prompt [y/N] " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]]
}

# Function to show file and line to fix
show_fix() {
    local file=$1
    local line=$2
    local issue=$3
    
    echo -e "\n${YELLOW}Issue:${NC} $issue"
    echo -e "${BLUE}File:${NC} $file"
    if [[ -n "$line" ]]; then
        echo -e "${BLUE}Line:${NC} $line"
        echo -e "\n${YELLOW}Current code:${NC}"
        sed -n "$((line-2)),$((line+2))p" "$file" 2>/dev/null || echo "File not found"
    fi
}

echo -e "${YELLOW}=== 1. CLI Fixes ===${NC}"

# Fix 1: Pipeline deployment endpoint
echo -e "\n${RED}Issue 1:${NC} CLI using wrong pipeline deployment endpoint"
show_fix "projects/phoenix-cli/internal/client/api.go" "" "DeployPipeline using /pipelines/deployments instead of /deployments"
echo -e "\n${GREEN}Fix:${NC} Change endpoint from '/api/v1/pipelines/deployments' to '/api/v1/deployments'"

if confirm "Open file to fix this?"; then
    ${EDITOR:-vim} projects/phoenix-cli/internal/client/api.go
fi

# Fix 2: Experiment metrics endpoint
echo -e "\n${RED}Issue 2:${NC} CLI experiment metrics using non-existent endpoint"
show_fix "projects/phoenix-cli/cmd/experiment_metrics.go" "" "Using /experiments/{id}/metrics instead of /kpis"
echo -e "\n${GREEN}Fix:${NC} Either:"
echo "  1. Update CLI to call both /kpis and /cost-analysis endpoints"
echo "  2. Add a unified /metrics endpoint to the API"

if confirm "Open file to fix this?"; then
    ${EDITOR:-vim} projects/phoenix-cli/cmd/experiment_metrics.go
fi

echo -e "\n${YELLOW}=== 2. API Fixes ===${NC}"

# Fix 3: Missing experiment metrics endpoint
echo -e "\n${RED}Issue 3:${NC} API missing unified experiment metrics endpoint"
show_fix "projects/phoenix-api/internal/api/experiments.go" "" "No handleGetExperimentMetrics handler"
echo -e "\n${GREEN}Fix:${NC} Add handler that combines KPI and cost data"

if confirm "Create the handler?"; then
    cat > /tmp/metrics_handler.go << 'EOF'
// Add to experiments.go
func (s *Server) handleGetExperimentMetrics(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    experimentID := chi.URLParam(r, "id")
    
    // Get KPIs
    kpis, err := s.metricsCollector.GetExperimentKPIs(ctx, experimentID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to get KPIs")
        return
    }
    
    // Get cost analysis
    costAnalysis, err := s.costService.CalculateExperimentCostSavings(ctx, experimentID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to get cost analysis")
        return
    }
    
    // Combine results
    metrics := map[string]interface{}{
        "experiment_id": experimentID,
        "kpis": kpis,
        "cost_analysis": costAnalysis,
        "timestamp": time.Now(),
    }
    
    respondJSON(w, http.StatusOK, metrics)
}
EOF
    echo -e "\n${GREEN}Sample handler created at /tmp/metrics_handler.go${NC}"
fi

# Fix 4: WebSocket events
echo -e "\n${RED}Issue 4:${NC} Missing WebSocket events for experiment lifecycle"
show_fix "projects/phoenix-api/internal/controller/experiment_controller.go" "" "No experiment_started broadcast"
echo -e "\n${GREEN}Fix:${NC} Add hub.Broadcast() calls for:"
echo "  - experiment_started"
echo "  - experiment_completed"
echo "  - experiment_analyzed"

echo -e "\n${YELLOW}=== 3. Agent Fixes ===${NC}"

# Fix 5: Rollback task handling
echo -e "\n${RED}Issue 5:${NC} Agent doesn't handle rollback action for deployments"
show_fix "projects/phoenix-agent/internal/supervisor/supervisor.go" "" "Missing case for Action:'rollback'"
echo -e "\n${GREEN}Fix:${NC} Add rollback case that stops the deployment"

if confirm "Show the fix code?"; then
    cat << 'EOF'
case "rollback":
    log.Info().Str("deployment_id", task.TargetID).Msg("Rolling back deployment")
    err := s.collectorMgr.Stop(task.TargetID)
    if err != nil {
        return fmt.Errorf("rollback failed: %w", err)
    }
    task.Status = "completed"
    task.Result = map[string]interface{}{
        "message": "Pipeline rolled back successfully",
    }
EOF
fi

echo -e "\n${YELLOW}=== 4. Pipeline Fixes ===${NC}"

# Fix 6: Template validation
echo -e "\n${RED}Issue 6:${NC} Pipeline validation endpoint is a stub"
show_fix "projects/phoenix-api/internal/api/pipelines.go" "" "handleValidatePipeline returns stub response"
echo -e "\n${GREEN}Fix:${NC} Implement actual YAML validation using OTel collector dry-run"

echo -e "\n${YELLOW}=== 5. Metrics Engine Fixes ===${NC}"

# Fix 7: Cost calculations
echo -e "\n${RED}Issue 7:${NC} CostService using placeholder calculations"
show_fix "projects/phoenix-api/internal/services/cost_service.go" "" "TODO: Implement GetExperimentMetrics"
echo -e "\n${GREEN}Fix:${NC} Use real metrics from store instead of estimates"

echo -e "\n${YELLOW}=== Quick Commands ===${NC}"
echo ""
echo "# To find all TODOs in the codebase:"
echo -e "${BLUE}grep -r \"TODO\" projects/ --include=\"*.go\" | grep -v vendor${NC}"
echo ""
echo "# To test a specific fix:"
echo -e "${BLUE}go test ./projects/phoenix-api/internal/api -run TestExperimentMetrics${NC}"
echo ""
echo "# To run the validation script:"
echo -e "${BLUE}./scripts/mvp-validation.sh${NC}"
echo ""
echo "# To check WebSocket events:"
echo -e "${BLUE}wscat -c ws://localhost:8080/api/v1/ws${NC}"

echo -e "\n${YELLOW}=== Next Steps ===${NC}"
echo "1. Fix CLI endpoint mismatches (High Priority)"
echo "2. Implement missing API handlers"
echo "3. Add WebSocket event broadcasts"
echo "4. Complete agent rollback handling"
echo "5. Run validation script after each fix"

echo -e "\n${GREEN}Good luck with the fixes! Run ./scripts/mvp-validation.sh to test progress.${NC}"