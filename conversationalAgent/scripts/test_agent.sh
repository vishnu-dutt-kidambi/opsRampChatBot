#!/bin/bash
# =============================================================================
# HPE Autopilot — Automated Test Suite
# =============================================================================
# Tests all agent flows via the HTTP API and validates responses.
#
# Usage:
#   1. Start the agent:  make web
#   2. Run tests:        ./test_agent.sh
#
# Each test clears conversation history, sends a query, and checks that the
# response contains expected keywords (case-insensitive).
# =============================================================================

BASE="${AGENT_URL:-http://localhost:8080}"
PASS=0
FAIL=0
TOTAL=0
TIMEOUT=180

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

# Check if server is running
if ! curl -sf "$BASE" > /dev/null 2>&1; then
    echo "❌ Agent not running at $BASE"
    echo "   Start it with: make web"
    exit 1
fi

test_query() {
    local name="$1"
    local query="$2"
    shift 2
    local expected_terms=("$@")

    ((TOTAL++))

    # Clear history between tests
    curl -s -X POST "$BASE/api/clear" > /dev/null 2>&1

    # Send query
    local response
    response=$(curl -s --max-time "$TIMEOUT" "$BASE/api/chat" \
        -H 'Content-Type: application/json' \
        -d "{\"message\":\"$query\"}" 2>/dev/null)

    if [[ -z "$response" ]]; then
        echo -e "${RED}❌ FAIL${NC}: $name — no response (timeout after ${TIMEOUT}s)"
        ((FAIL++))
        return
    fi

    local answer
    answer=$(echo "$response" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(data.get('answer', ''))
except:
    print('')
" 2>/dev/null)

    if [[ -z "$answer" ]]; then
        echo -e "${RED}❌ FAIL${NC}: $name — empty or invalid answer"
        echo "   Raw: ${response:0:200}"
        ((FAIL++))
        return
    fi

    # Check all expected terms
    local all_found=true
    local missing=""
    for term in "${expected_terms[@]}"; do
        if ! echo "$answer" | grep -qi "$term"; then
            all_found=false
            missing="$missing '$term'"
        fi
    done

    if $all_found; then
        echo -e "${GREEN}✅ PASS${NC}: $name"
        ((PASS++))
    else
        echo -e "${RED}❌ FAIL${NC}: $name — missing:$missing"
        echo "   Answer (first 300 chars): ${answer:0:300}"
        ((FAIL++))
    fi
}

echo ""
echo "🧪 HPE Autopilot — Test Suite"
echo "=============================================="
echo "Target: $BASE"
echo ""

# ─── Flow 1: Search Alerts ───────────────────────────────────────────────────
echo -e "${YELLOW}── Alerts ──${NC}"
test_query \
    "Critical alerts" \
    "Show me all critical alerts" \
    "web-server-prod-01" "db-primary-01"

test_query \
    "Warning alerts" \
    "Show me all warning alerts" \
    "Warning"

test_query \
    "Alerts by resource" \
    "Are there any alerts for db-primary-01?" \
    "db-primary-01"

# ─── Flow 2: Search Resources ────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}── Resources ──${NC}"
test_query \
    "AWS resources" \
    "List all resources running in AWS" \
    "web-server-prod-01" "db-primary-01"

test_query \
    "GCP resources" \
    "Show me all GCP resources" \
    "GCP"

# ─── Flow 3: Resource Details ────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}── Resource Details ──${NC}"
test_query \
    "DB details" \
    "Show me details of db-primary-01" \
    "10.0.2.10" "r6g.4xlarge"

# ─── Flow 4: Search Incidents ────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}── Incidents ──${NC}"
test_query \
    "Open incidents" \
    "Show me all open incidents" \
    "INC-20260219-001" "INC-20260219-002"

test_query \
    "Urgent incidents" \
    "Show me urgent incidents" \
    "Urgent"

# ─── Flow 5: Investigate Resource ────────────────────────────────────────────
echo ""
echo -e "${YELLOW}── Investigation ──${NC}"
test_query \
    "Investigate web server" \
    "Investigate web-server-prod-01" \
    "97.3" "ALR-20260219-001"

# ─── Flow 6: Environment Summary ─────────────────────────────────────────────
echo ""
echo -e "${YELLOW}── Environment Summary ──${NC}"
test_query \
    "Environment summary" \
    "Give me an environment summary" \
    "22" "AWS"

# ─── Flow 7: Capacity Forecast (Single) ──────────────────────────────────────
echo ""
echo -e "${YELLOW}── Capacity Forecasting ──${NC}"
test_query \
    "Forecast single resource" \
    "Predict capacity for web-server-prod-01" \
    "CPU" "Memory" "Disk"

test_query \
    "Forecast disk for DB" \
    "When will db-primary-01 run out of disk?" \
    "db-primary-01" "disk"

# ─── Flow 8: Capacity Forecast (All Resources) ──────────────────────────────
test_query \
    "All at-risk resources" \
    "Which resources are at risk of running out of capacity?" \
    "web-server-prod-01" "db-primary-01"

# ─── Flow 9: Knowledge Base (RAG) ───────────────────────────────────────────
echo ""
echo -e "${YELLOW}── Knowledge Base ──${NC}"
test_query \
    "High CPU runbook" \
    "What is the runbook for high CPU usage?" \
    "CPU" "top"

test_query \
    "Disk full procedure" \
    "How do I troubleshoot disk space full?" \
    "disk" "df"

test_query \
    "Escalation contacts" \
    "What are the escalation contacts?" \
    "escalat"

# ─── Results ─────────────────────────────────────────────────────────────────
echo ""
echo "=============================================="
if [[ $FAIL -eq 0 ]]; then
    echo -e "${GREEN}All $TOTAL tests passed!${NC}"
else
    echo -e "Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC} out of $TOTAL tests"
fi
echo ""
exit $FAIL
