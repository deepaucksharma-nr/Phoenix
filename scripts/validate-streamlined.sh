#!/bin/bash
set -e

echo "🔍 Phoenix-vNext Streamlined Structure Validation"
echo "================================================"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Validation counters
ERRORS=0
WARNINGS=0

# Function to check file exists
check_file() {
    local file=$1
    local description=$2
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $description exists"
    else
        echo -e "${RED}✗${NC} $description missing: $file"
        ((ERRORS++))
    fi
}

# Function to check directory exists
check_dir() {
    local dir=$1
    local description=$2
    if [ -d "$dir" ]; then
        echo -e "${GREEN}✓${NC} $description exists"
    else
        echo -e "${RED}✗${NC} $description missing: $dir"
        ((ERRORS++))
    fi
}

# Function to check for duplicates
check_no_duplicates() {
    local pattern=$1
    local description=$2
    local count=$(find . -name "$pattern" -not -path "./archive/*" 2>/dev/null | wc -l)
    if [ "$count" -le 1 ]; then
        echo -e "${GREEN}✓${NC} No duplicate $description"
    else
        echo -e "${YELLOW}⚠${NC} Found $count instances of $description"
        find . -name "$pattern" -not -path "./archive/*" 2>/dev/null | head -5
        ((WARNINGS++))
    fi
}

echo -e "\n📁 Checking Core Structure..."
check_dir "configs" "Configs directory"
check_dir "configs/monitoring/prometheus" "Prometheus configs"
check_dir "configs/monitoring/grafana/dashboards" "Grafana dashboards"
check_dir "configs/otel/collectors" "OTel collectors"
check_dir "k8s/base" "Kubernetes base"
check_dir "runbooks" "Runbooks"
check_dir "services" "Services"
check_dir "apps" "Apps"

echo -e "\n📄 Checking Essential Files..."
check_file "docker-compose.yaml" "Main docker-compose"
check_file "configs/monitoring/prometheus/prometheus.yaml" "Prometheus config"
check_file "configs/monitoring/prometheus/rules/phoenix_rules.yml" "Consolidated rules"
check_file "configs/otel/collectors/main.yaml" "Main collector config"
check_file "configs/otel/collectors/observer.yaml" "Observer collector config"

echo -e "\n🔍 Checking for Duplicates..."
check_no_duplicates "prometheus.yaml" "Prometheus configs"
check_no_duplicates "phoenix_rules*.yml" "Prometheus rules"
check_no_duplicates "docker-compose*.yaml" "Docker compose files"

echo -e "\n📊 Dashboard Count..."
DASHBOARD_COUNT=$(ls configs/monitoring/grafana/dashboards/*.json 2>/dev/null | wc -l)
echo "Found $DASHBOARD_COUNT dashboards in canonical location"

echo -e "\n🏗️ Service Structure..."
# Check service consistency
COMPOSE_SERVICES=$(grep -E "^  [a-zA-Z-]+:$" docker-compose.yaml | wc -l)
echo "Docker Compose defines $COMPOSE_SERVICES services"

echo -e "\n✨ Streamlining Benefits..."
# Calculate space saved
if [ -d "archive" ]; then
    ARCHIVE_SIZE=$(du -sh archive 2>/dev/null | cut -f1)
    echo "Archived redundant files: $ARCHIVE_SIZE"
fi

# Check for old paths in docker-compose
echo -e "\n🔗 Checking Path References..."
OLD_PATHS=("monitoring/prometheus" "config/defaults" "infrastructure/kubernetes")
for path in "${OLD_PATHS[@]}"; do
    if grep -q "$path" docker-compose.yaml 2>/dev/null; then
        echo -e "${YELLOW}⚠${NC} Found reference to old path: $path"
        ((WARNINGS++))
    fi
done

echo -e "\n📈 Summary"
echo "========="
echo -e "Errors: ${ERRORS}"
echo -e "Warnings: ${WARNINGS}"

if [ $ERRORS -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo -e "\n${GREEN}✅ Phoenix-vNext is properly streamlined!${NC}"
        exit 0
    else
        echo -e "\n${YELLOW}⚠️  Streamlining complete with warnings${NC}"
        exit 0
    fi
else
    echo -e "\n${RED}❌ Streamlining validation failed${NC}"
    exit 1
fi