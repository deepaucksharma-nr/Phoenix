#!/bin/bash
# Phoenix Platform Performance Benchmark Script
# Comprehensive performance testing for Phoenix API and CLI

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
API_URL=${PHOENIX_API_URL:-"http://localhost:8080"}
CLI_PATH=${PHOENIX_CLI_PATH:-"$PROJECT_ROOT/bin/phoenix"}
RESULTS_DIR=${BENCHMARK_RESULTS_DIR:-"$PROJECT_ROOT/benchmark-results"}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$RESULTS_DIR/benchmark_report_$TIMESTAMP.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configurations
declare -A API_TESTS=(
    ["list_experiments"]="GET /api/v1/experiments 100 10"
    ["list_deployments"]="GET /api/v1/pipeline-deployments 50 5"
    ["health_check"]="GET /health 200 20"
)

declare -A LOAD_TESTS=(
    ["light_load"]="30s 5 constant"
    ["medium_load"]="60s 15 constant"
    ["heavy_load"]="120s 50 constant"
    ["spike_test"]="60s 100 spike"
)

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
Phoenix Platform Benchmark Suite

USAGE:
    $0 [OPTIONS]

OPTIONS:
    --api-url URL          Phoenix API URL (default: http://localhost:8080)
    --cli-path PATH        Path to Phoenix CLI binary
    --results-dir PATH     Directory to store results (default: ./benchmark-results)
    --quick               Run quick benchmark (reduced load)
    --load-only           Run only load tests
    --api-only            Run only API tests
    --cleanup             Clean up test data after benchmarks
    --verbose             Enable verbose output
    --help                Show this help message

EXAMPLES:
    # Run full benchmark suite
    $0

    # Quick benchmark for CI/CD
    $0 --quick

    # Test specific API endpoint
    $0 --api-only --api-url https://staging.phoenix.example.com

    # Load testing only
    $0 --load-only --results-dir /tmp/load-results

ENVIRONMENT VARIABLES:
    PHOENIX_API_URL          API endpoint URL
    PHOENIX_CLI_PATH         Path to CLI binary
    BENCHMARK_RESULTS_DIR    Results directory
    PHOENIX_API_TOKEN        API authentication token
EOF
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if API is accessible
    if ! curl -s "$API_URL/health" > /dev/null 2>&1; then
        log_error "Phoenix API is not accessible at $API_URL"
        log_error "Please ensure the API server is running"
        exit 1
    fi
    
    # Check if CLI is available
    if [[ ! -f "$CLI_PATH" ]]; then
        log_warning "CLI not found at $CLI_PATH, trying to build it..."
        cd "$PROJECT_ROOT"
        make build-cli || {
            log_error "Failed to build CLI"
            exit 1
        }
        CLI_PATH="$PROJECT_ROOT/bin/phoenix"
    fi
    
    # Check CLI authentication
    if ! "$CLI_PATH" auth status > /dev/null 2>&1; then
        log_warning "CLI not authenticated, attempting login..."
        if [[ -n "$PHOENIX_API_TOKEN" ]]; then
            export PHOENIX_AUTH_TOKEN="$PHOENIX_API_TOKEN"
        else
            log_error "No authentication available. Please set PHOENIX_API_TOKEN or run 'phoenix auth login'"
            exit 1
        fi
    fi
    
    # Check required tools
    for tool in jq curl bc; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "Required tool '$tool' is not installed"
            exit 1
        fi
    done
    
    # Create results directory
    mkdir -p "$RESULTS_DIR"
    
    log_success "Prerequisites check passed"
}

# System resource monitoring
start_monitoring() {
    log_info "Starting system resource monitoring..."
    
    MONITOR_PID=""
    if command -v top &> /dev/null; then
        # Start resource monitoring in background
        {
            while true; do
                echo "$(date '+%Y-%m-%d %H:%M:%S'),$(top -l 1 -n 0 | grep "CPU usage" | cut -d: -f2 | cut -d% -f1 | tr -d ' ')" >> "$RESULTS_DIR/cpu_usage_$TIMESTAMP.csv"
                sleep 5
            done
        } &
        MONITOR_PID=$!
    fi
}

stop_monitoring() {
    if [[ -n "$MONITOR_PID" ]]; then
        kill "$MONITOR_PID" 2>/dev/null || true
        log_info "Stopped system monitoring"
    fi
}

# API endpoint benchmarks
run_api_benchmarks() {
    log_info "Running API endpoint benchmarks..."
    
    local results=()
    
    for test_name in "${!API_TESTS[@]}"; do
        IFS=' ' read -ra test_config <<< "${API_TESTS[$test_name]}"
        local method="${test_config[0]}"
        local endpoint="${test_config[1]}"
        local requests="${test_config[2]}"
        local concurrency="${test_config[3]}"
        
        # Adjust for quick mode
        if [[ "$QUICK_MODE" == "true" ]]; then
            requests=$((requests / 5))
            concurrency=$((concurrency / 2))
            [[ $requests -lt 10 ]] && requests=10
            [[ $concurrency -lt 2 ]] && concurrency=2
        fi
        
        log_info "Testing $test_name ($method $endpoint)..."
        
        local result
        result=$("$CLI_PATH" benchmark api \
            --endpoint "$endpoint" \
            --method "$method" \
            --requests "$requests" \
            --concurrency "$concurrency" \
            --output json 2>/dev/null)
        
        if [[ $? -eq 0 ]]; then
            # Add test metadata
            result=$(echo "$result" | jq ". + {\"test_name\": \"$test_name\", \"endpoint\": \"$endpoint\", \"method\": \"$method\"}")
            results+=("$result")
            log_success "Completed $test_name"
        else
            log_error "Failed $test_name"
            results+=("{\"test_name\": \"$test_name\", \"error\": \"benchmark_failed\"}")
        fi
    done
    
    # Combine results
    local combined_results="["
    for i in "${!results[@]}"; do
        [[ $i -gt 0 ]] && combined_results+=","
        combined_results+="${results[$i]}"
    done
    combined_results+="]"
    
    echo "$combined_results" > "$RESULTS_DIR/api_benchmarks_$TIMESTAMP.json"
    log_success "API benchmarks completed"
}

# Load testing
run_load_tests() {
    log_info "Running load tests..."
    
    local results=()
    
    for test_name in "${!LOAD_TESTS[@]}"; do
        IFS=' ' read -ra test_config <<< "${LOAD_TESTS[$test_name]}"
        local duration="${test_config[0]}"
        local rps="${test_config[1]}"
        local pattern="${test_config[2]}"
        
        # Adjust for quick mode
        if [[ "$QUICK_MODE" == "true" ]]; then
            duration="10s"
            rps=$((rps / 2))
            [[ $rps -lt 2 ]] && rps=2
        fi
        
        log_info "Running $test_name load test (${duration}, ${rps} RPS, ${pattern})..."
        
        local result
        result=$("$CLI_PATH" benchmark load \
            --duration "$duration" \
            --rps "$rps" \
            --pattern "$pattern" \
            --endpoints "/api/v1/experiments" \
            --output json 2>/dev/null)
        
        if [[ $? -eq 0 ]]; then
            result=$(echo "$result" | jq ". + {\"test_name\": \"$test_name\", \"duration\": \"$duration\", \"target_rps\": $rps, \"pattern\": \"$pattern\"}")
            results+=("$result")
            log_success "Completed $test_name"
        else
            log_error "Failed $test_name"
            results+=("{\"test_name\": \"$test_name\", \"error\": \"load_test_failed\"}")
        fi
        
        # Brief pause between load tests
        sleep 5
    done
    
    # Combine results
    local combined_results="["
    for i in "${!results[@]}"; do
        [[ $i -gt 0 ]] && combined_results+=","
        combined_results+="${results[$i]}"
    done
    combined_results+="]"
    
    echo "$combined_results" > "$RESULTS_DIR/load_tests_$TIMESTAMP.json"
    log_success "Load tests completed"
}

# Experiment operations benchmark
run_experiment_benchmarks() {
    log_info "Running experiment operation benchmarks..."
    
    local experiments=20
    local concurrency=5
    
    if [[ "$QUICK_MODE" == "true" ]]; then
        experiments=5
        concurrency=2
    fi
    
    log_info "Testing experiment operations ($experiments experiments, $concurrency concurrency)..."
    
    local result
    result=$("$CLI_PATH" benchmark experiment \
        --experiments "$experiments" \
        --concurrency "$concurrency" \
        --cleanup \
        --output json 2>/dev/null)
    
    if [[ $? -eq 0 ]]; then
        echo "$result" > "$RESULTS_DIR/experiment_benchmarks_$TIMESTAMP.json"
        log_success "Experiment benchmarks completed"
    else
        log_error "Experiment benchmarks failed"
        echo '{"error": "experiment_benchmark_failed"}' > "$RESULTS_DIR/experiment_benchmarks_$TIMESTAMP.json"
    fi
}

# Memory and resource profiling
run_resource_profiling() {
    log_info "Running resource profiling..."
    
    # Profile API server memory usage during load
    {
        local duration=60
        [[ "$QUICK_MODE" == "true" ]] && duration=30
        
        log_info "Profiling API server resources for ${duration}s..."
        
        # Start background load
        "$CLI_PATH" benchmark load \
            --duration "${duration}s" \
            --rps 20 \
            --pattern constant \
            --endpoints "/api/v1/experiments" > /dev/null 2>&1 &
        local load_pid=$!
        
        # Monitor resources
        local start_time=$(date +%s)
        while [[ $(($(date +%s) - start_time)) -lt $duration ]]; do
            # Get API server process info (adjust based on your setup)
            local memory_mb=0
            local cpu_percent=0
            
            # Try to find Phoenix API process
            if pgrep -f "phoenix.*api" > /dev/null; then
                local pid=$(pgrep -f "phoenix.*api" | head -1)
                if [[ -n "$pid" ]]; then
                    # Get memory usage (in MB)
                    memory_mb=$(ps -o rss= -p "$pid" 2>/dev/null | awk '{print $1/1024}' || echo "0")
                    # Get CPU percentage
                    cpu_percent=$(ps -o %cpu= -p "$pid" 2>/dev/null | tr -d ' ' || echo "0")
                fi
            fi
            
            echo "$(date '+%Y-%m-%d %H:%M:%S'),$memory_mb,$cpu_percent" >> "$RESULTS_DIR/resource_profile_$TIMESTAMP.csv"
            sleep 2
        done
        
        # Wait for load test to finish
        wait $load_pid 2>/dev/null || true
        
    } || log_warning "Resource profiling failed"
    
    log_success "Resource profiling completed"
}

# Generate comprehensive report
generate_report() {
    log_info "Generating comprehensive benchmark report..."
    
    local report_data="{
        \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
        \"test_environment\": {
            \"api_url\": \"$API_URL\",
            \"cli_version\": \"$("$CLI_PATH" version --short 2>/dev/null || echo "unknown")\",
            \"system\": \"$(uname -s)\",
            \"arch\": \"$(uname -m)\",
            \"quick_mode\": $([[ "$QUICK_MODE" == "true" ]] && echo "true" || echo "false")
        },
        \"results\": {}
    }"
    
    # Add API benchmark results
    if [[ -f "$RESULTS_DIR/api_benchmarks_$TIMESTAMP.json" ]]; then
        local api_results
        api_results=$(cat "$RESULTS_DIR/api_benchmarks_$TIMESTAMP.json")
        report_data=$(echo "$report_data" | jq ".results.api_benchmarks = $api_results")
    fi
    
    # Add load test results
    if [[ -f "$RESULTS_DIR/load_tests_$TIMESTAMP.json" ]]; then
        local load_results
        load_results=$(cat "$RESULTS_DIR/load_tests_$TIMESTAMP.json")
        report_data=$(echo "$report_data" | jq ".results.load_tests = $load_results")
    fi
    
    # Add experiment benchmark results
    if [[ -f "$RESULTS_DIR/experiment_benchmarks_$TIMESTAMP.json" ]]; then
        local exp_results
        exp_results=$(cat "$RESULTS_DIR/experiment_benchmarks_$TIMESTAMP.json")
        report_data=$(echo "$report_data" | jq ".results.experiment_benchmarks = $exp_results")
    fi
    
    # Add resource profiling data
    if [[ -f "$RESULTS_DIR/resource_profile_$TIMESTAMP.csv" ]]; then
        local resource_summary="{
            \"file\": \"resource_profile_$TIMESTAMP.csv\",
            \"description\": \"Memory and CPU usage during load testing\"
        }"
        report_data=$(echo "$report_data" | jq ".results.resource_profiling = $resource_summary")
    fi
    
    echo "$report_data" > "$REPORT_FILE"
    
    # Generate human-readable summary
    generate_summary_report
    
    log_success "Benchmark report generated: $REPORT_FILE"
}

# Generate human-readable summary
generate_summary_report() {
    local summary_file="$RESULTS_DIR/benchmark_summary_$TIMESTAMP.txt"
    
    cat > "$summary_file" << EOF
Phoenix Platform Benchmark Report
Generated: $(date)
Test Environment: $API_URL
Quick Mode: $([[ "$QUICK_MODE" == "true" ]] && echo "Yes" || echo "No")

=== API ENDPOINT BENCHMARKS ===
EOF
    
    if [[ -f "$RESULTS_DIR/api_benchmarks_$TIMESTAMP.json" ]]; then
        echo "$(cat "$RESULTS_DIR/api_benchmarks_$TIMESTAMP.json" | jq -r '.[] | 
            "Test: " + .test_name + 
            "\n  Requests/sec: " + (.requests_per_second | tostring) + 
            "\n  Avg Latency: " + (.avg_latency | tostring) + 
            "\n  Error Rate: " + (.error_rate | tostring) + "%\n"')" >> "$summary_file"
    else
        echo "No API benchmark data available" >> "$summary_file"
    fi
    
    cat >> "$summary_file" << EOF

=== LOAD TEST RESULTS ===
EOF
    
    if [[ -f "$RESULTS_DIR/load_tests_$TIMESTAMP.json" ]]; then
        echo "$(cat "$RESULTS_DIR/load_tests_$TIMESTAMP.json" | jq -r '.[] | 
            "Test: " + .test_name + 
            "\n  Achieved RPS: " + (.requests_per_second | tostring) + 
            "\n  Avg Latency: " + (.avg_latency | tostring) + 
            "\n  P95 Latency: " + (.p95_latency | tostring) + 
            "\n  Error Rate: " + (.error_rate | tostring) + "%\n"')" >> "$summary_file"
    else
        echo "No load test data available" >> "$summary_file"
    fi
    
    cat >> "$summary_file" << EOF

=== FILES GENERATED ===
- Full Report: $(basename "$REPORT_FILE")
- Summary: $(basename "$summary_file")
EOF
    
    if [[ -f "$RESULTS_DIR/resource_profile_$TIMESTAMP.csv" ]]; then
        echo "- Resource Profile: resource_profile_$TIMESTAMP.csv" >> "$summary_file"
    fi
    
    if [[ -f "$RESULTS_DIR/cpu_usage_$TIMESTAMP.csv" ]]; then
        echo "- CPU Usage: cpu_usage_$TIMESTAMP.csv" >> "$summary_file"
    fi
    
    log_info "Summary report: $summary_file"
}

# Cleanup function
cleanup() {
    stop_monitoring
    
    if [[ "$CLEANUP_MODE" == "true" ]]; then
        log_info "Cleaning up test data..."
        # Add cleanup logic here if needed
    fi
    
    log_info "Benchmark run completed"
}

# Main execution
main() {
    # Parse command line arguments
    QUICK_MODE=false
    LOAD_ONLY=false
    API_ONLY=false
    CLEANUP_MODE=false
    VERBOSE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --api-url)
                API_URL="$2"
                shift 2
                ;;
            --cli-path)
                CLI_PATH="$2"
                shift 2
                ;;
            --results-dir)
                RESULTS_DIR="$2"
                REPORT_FILE="$RESULTS_DIR/benchmark_report_$TIMESTAMP.json"
                shift 2
                ;;
            --quick)
                QUICK_MODE=true
                shift
                ;;
            --load-only)
                LOAD_ONLY=true
                shift
                ;;
            --api-only)
                API_ONLY=true
                shift
                ;;
            --cleanup)
                CLEANUP_MODE=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                set -x
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo
                show_help
                exit 1
                ;;
        esac
    done
    
    # Set up trap for cleanup
    trap cleanup EXIT
    
    log_info "Starting Phoenix Platform benchmark suite..."
    log_info "API URL: $API_URL"
    log_info "Results directory: $RESULTS_DIR"
    
    check_prerequisites
    start_monitoring
    
    # Run benchmarks based on flags
    if [[ "$API_ONLY" == "true" ]]; then
        run_api_benchmarks
    elif [[ "$LOAD_ONLY" == "true" ]]; then
        run_load_tests
    else
        # Run full suite
        run_api_benchmarks
        run_load_tests
        run_experiment_benchmarks
        run_resource_profiling
    fi
    
    generate_report
    
    log_success "Benchmark suite completed successfully!"
    log_info "Results available in: $RESULTS_DIR"
}

# Run main function
main "$@"