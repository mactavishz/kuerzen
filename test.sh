#!/bin/bash

# Kuerzen Test Runner
# Runs tests for all modules in the workspace

echo "🧪 Running tests for all modules in the workspace..."
echo "=================================================="

# Define modules in workspace
modules=(
    "analytics"
    "middleware"
    "redirector"
    "retries"
    "shortener"
    "store"
)

success=true
failed_modules=()

for module in "${modules[@]}"; do
    echo ""
    echo "📦 Testing module: $module"
    echo "-----------------------------------"

    if go test "./$module/..." -v; then
        echo "✅ $module: PASSED"
    else
        echo "❌ $module: FAILED"
        success=false
        failed_modules+=("$module")
    fi
done

echo ""
echo "=================================================="
if [ "$success" = true ]; then
    echo "🎉 All tests passed!"
else
    echo "❌ Some tests failed:"
    for module in "${failed_modules[@]}"; do
        echo "  - $module"
    done
    exit 1
fi
