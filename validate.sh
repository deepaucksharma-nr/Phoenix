#!/bin/bash

# Phoenix Platform Migration Validation
# Quick validation script to check migration status

echo "🔍 Validating Phoenix Platform Migration..."
echo ""

# Check if we're in the right directory
if [ ! -f "go.work" ]; then
    echo "❌ Not in Phoenix root directory"
    exit 1
fi

echo "✅ In Phoenix Platform directory"

# Check go.work exists and has key modules
if grep -q "packages/go-common" go.work && grep -q "services/phoenix-cli" go.work; then
    echo "✅ Go workspace properly configured"
else
    echo "⚠️  Go workspace may need updates"
fi

# Check if Phoenix CLI exists
if [ -d "services/phoenix-cli" ]; then
    echo "✅ Phoenix CLI service exists"
    if [ -f "services/phoenix-cli/bin/phoenix" ]; then
        echo "✅ Phoenix CLI binary exists"
    else
        echo "📋 Phoenix CLI needs building (run: cd services/phoenix-cli && go build -o bin/phoenix .)"
    fi
else
    echo "❌ Phoenix CLI missing"
fi

# Check core services
for service in api controller generator; do
    if [ -d "services/$service" ]; then
        echo "✅ $service service exists"
    else
        echo "❌ $service service missing"
    fi
done

# Check packages
if [ -d "packages/go-common" ] && [ -d "packages/contracts" ]; then
    echo "✅ Shared packages exist"
else
    echo "❌ Shared packages missing"
fi

# Check documentation
docs=("MIGRATION_COMPLETE.md" "DEVELOPMENT_GUIDE.md" "QUICK_START.md" "NEXT_STEPS.md")
for doc in "${docs[@]}"; do
    if [ -f "$doc" ]; then
        echo "✅ $doc exists"
    else
        echo "❌ $doc missing"
    fi
done

echo ""
echo "🎉 Migration validation complete!"
echo ""
echo "Next steps:"
echo "1. Install protoc: bash scripts/install-protoc.sh"
echo "2. Generate protos: cd packages/contracts && bash generate.sh"
echo "3. Build CLI: cd services/phoenix-cli && go build -o bin/phoenix ."
echo "4. Build all: go work sync && make build-all"