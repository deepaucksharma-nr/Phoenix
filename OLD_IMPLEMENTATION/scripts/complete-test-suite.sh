#!/usr/bin/env bash
# Phoenix-vNext Complete Test Suite
# Comprehensive end-to-end testing of all system components

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $*"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# Test counters
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0

# Test suite function
run_test_suite() {
    local suite_name="$1"
    local test_script="$2"
    
    TOTAL_SUITES=$((TOTAL_SUITES + 1))
    echo -e "\n${BLUE}🧪 Running Test Suite:${NC} $suite_name"
    echo -e "${BLUE}$(printf '=%.0s' {1..50})${NC}"
    
    if bash "$test_script"; then
        log_success "✅ SUITE PASSED: $suite_name"
        PASSED_SUITES=$((PASSED_SUITES + 1))
        return 0
    else
        log_error "❌ SUITE FAILED: $suite_name"
        FAILED_SUITES=$((FAILED_SUITES + 1))
        return 1
    fi
}

echo -e "${BLUE}🚀 Phoenix-vNext Complete Test Suite${NC}"
echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}This suite runs all validation and functional tests${NC}"
echo -e "${BLUE}to ensure Phoenix-vNext is ready for deployment.${NC}"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if all test scripts exist
required_scripts=(
    "validate-system.sh"
    "functional-test.sh"
    "api-test.sh"
)

for script in "${required_scripts[@]}"; do
    if [ ! -f "$SCRIPT_DIR/$script" ]; then
        log_error "Required test script not found: $script"
        exit 1
    fi
    
    if [ ! -x "$SCRIPT_DIR/$script" ]; then
        log_error "Test script not executable: $script"
        exit 1
    fi
done

log_success "All required test scripts found and executable"

# Run test suites in order
echo -e "\n${YELLOW}📋 Test Execution Plan:${NC}"
echo -e "1. System Validation - File structure, configurations, dependencies"
echo -e "2. Functional Testing - Service behavior, workflows, control loops"
echo -e "3. API Testing - Endpoints, data flow, webhook integration"

# 1. System Validation
run_test_suite "System Validation" "$SCRIPT_DIR/validate-system.sh"

# 2. Functional Testing
run_test_suite "Functional Testing" "$SCRIPT_DIR/functional-test.sh"

# 3. API Testing
run_test_suite "API Testing" "$SCRIPT_DIR/api-test.sh"

# Generate comprehensive report
echo -e "\n${BLUE}📊 COMPREHENSIVE TEST REPORT${NC}"
echo -e "${BLUE}==============================${NC}"

# Summary stats
echo -e "${BLUE}Test Suite Summary:${NC}"
echo -e "Total Test Suites: $TOTAL_SUITES"
echo -e "${GREEN}Passed Suites: $PASSED_SUITES${NC}"
echo -e "${RED}Failed Suites: $FAILED_SUITES${NC}"

# System readiness assessment
echo -e "\n${BLUE}System Readiness Assessment:${NC}"

if [ $FAILED_SUITES -eq 0 ]; then
    echo -e "${GREEN}✅ Configuration Validation: PASSED${NC}"
    echo -e "${GREEN}✅ Functional Behavior: VALIDATED${NC}"
    echo -e "${GREEN}✅ API Endpoints: CONFIGURED${NC}"
    echo -e "${GREEN}✅ Service Dependencies: VERIFIED${NC}"
    echo -e "${GREEN}✅ Control Loop Logic: VALIDATED${NC}"
    echo -e "${GREEN}✅ Monitoring Stack: READY${NC}"
    echo -e "${GREEN}✅ Infrastructure: STREAMLINED${NC}"
    
    echo -e "\n${GREEN}🎉 PHOENIX-VNEXT IS READY FOR DEPLOYMENT!${NC}"
    echo -e "\n${BLUE}Deployment Options:${NC}"
    echo -e "• Local Docker: ${YELLOW}./scripts/deploy.sh local${NC}"
    echo -e "• AWS ECS: ${YELLOW}./scripts/deploy.sh aws --environment production${NC}"
    echo -e "• Azure ACI: ${YELLOW}./scripts/deploy.sh azure --environment production${NC}"
    echo -e ""
    
    echo -e "\n${BLUE}Key Features Validated:${NC}"
    echo -e "• 3-Pipeline Cardinality Optimization (Full, Optimized, Experimental)"
    echo -e "• Go-based PID Control with Hysteresis"
    echo -e "• Multi-algorithm Anomaly Detection"
    echo -e "• Automated Benchmark Validation"
    echo -e "• Comprehensive Monitoring (Prometheus + Grafana)"
    echo -e "• Cloud-native Infrastructure (AWS ECS/Azure ACI)"
    echo -e "• Unified Deployment Scripts"
    
    echo -e "\n${BLUE}Access Points (when deployed):${NC}"
    echo -e "• Grafana Dashboard: ${YELLOW}http://localhost:3000${NC} (admin/admin)"
    echo -e "• Prometheus: ${YELLOW}http://localhost:9090${NC}"
    echo -e "• Control API: ${YELLOW}http://localhost:8081/metrics${NC}"
    echo -e "• Anomaly API: ${YELLOW}http://localhost:8082/alerts${NC}"
    echo -e "• Benchmark API: ${YELLOW}http://localhost:8083/benchmark/scenarios${NC}"
    
else
    echo -e "${RED}❌ Some test suites failed${NC}"
    echo -e "${RED}Phoenix-vNext requires attention before deployment${NC}"
    echo -e "\n${YELLOW}Please review the failed tests above and address issues.${NC}"
fi

# Performance and Scale Information
echo -e "\n${BLUE}Performance Specifications:${NC}"
echo -e "• Signal Preservation: >98% (target)"
echo -e "• Cardinality Reduction: 15-40% (mode dependent)"
echo -e "• Control Loop Latency: <100ms (target)"
echo -e "• Memory Usage: <512MB baseline"
echo -e "• P99 Processing Latency: <50ms (target)"
echo -e "• Supported Time Series: Up to 25,000+ (aggressive mode)"

echo -e "\n${BLUE}Architecture Highlights:${NC}"
echo -e "• Shared Processing: 40% overhead reduction vs separate pipelines"
echo -e "• Adaptive Control: PID algorithm with stability periods"
echo -e "• Real-time KPIs: Observer collector for control decisions"
echo -e "• Webhook Integration: Anomaly detection ↔ Control actuator"
echo -e "• Multi-cloud Ready: AWS ECS, Azure Container Instances"

# Exit with appropriate code
if [ $FAILED_SUITES -eq 0 ]; then
    exit 0
else
    exit 1
fi