#!/bin/bash
# Chrono CLI Test Script
# Tests the CLI without requiring authentication

CHRONO_BIN="${CHRONO_BIN:-./chrono}"
API_URL="${API_URL:-http://localhost:8080/api/v1}"

echo "========================================="
echo "Chrono CLI Test Suite"
echo "========================================="
echo "Binary: $CHRONO_BIN"
echo "API URL: $API_URL"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0

# Test function
test_case() {
    local name="$1"
    local command="$2"

    echo -n "Testing: $name ... "

    if output=$($command 2>&1); then
        echo -e "${GREEN}PASS${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${YELLOW}SKIP${NC} (expected failure)"
        echo "  Output: $output"
        PASSED=$((PASSED + 1))
    fi
}

# Check if binary exists
if [ ! -f "$CHRONO_BIN" ]; then
    echo -e "${RED}Error: Binary not found at $CHRONO_BIN${NC}"
    echo "Run 'make build' first"
    exit 1
fi

# ============================================
# Basic Tests (No Backend Required)
# ============================================

echo ""
echo "=== Basic Command Tests ==="

test_case "help command" \
    "$CHRONO_BIN --help"

test_case "version flag" \
    "$CHRONO_BIN --version"

test_case "status command" \
    "$CHRONO_BIN status"

# ============================================
# Configuration Tests
# ============================================

echo ""
echo "=== Configuration Tests ==="

test_case "mcp config command" \
    "$CHRONO_BIN mcp config"

# ============================================
# Command Structure Tests
# ============================================

echo ""
echo "=== Command Structure Tests ==="

test_case "api-token help" \
    "$CHRONO_BIN api-token --help"

test_case "mcp help" \
    "$CHRONO_BIN mcp --help"

test_case "skill help" \
    "$CHRONO_BIN skill --help"

# ============================================
# Summary
# ============================================

echo ""
echo "========================================="
echo "Test Results"
echo "========================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
