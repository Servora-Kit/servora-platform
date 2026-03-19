#!/usr/bin/env bash
# E2E test: IAM M2M Client Credentials Grant
# Tests the complete flow: create m2m app -> get token via client_credentials -> cleanup

set -euo pipefail

IAM_URL="${IAM_URL:-http://localhost:8000}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@servora.dev}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-changeme}"
APP_NAME="m2m-e2e-test-$(date +%s)"

PASS=0
FAIL=0

pass() { echo "[PASS] $1"; ((PASS++)) || true; }
fail() { echo "[FAIL] $1"; ((FAIL++)) || true; }

require_cmd() {
  if ! command -v "$1" &>/dev/null; then
    echo "ERROR: '$1' is required but not installed."
    exit 1
  fi
}

require_cmd curl
require_cmd jq

echo "=== IAM M2M Client Credentials E2E Test ==="
echo "Target: $IAM_URL"
echo ""

# Step 1: Admin login
echo "--- Step 1: Admin login ---"
LOGIN_RESP=$(curl -sf --max-time 10 -X POST "$IAM_URL/v1/auth/login/email-password" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}" 2>&1) || {
  fail "Admin login request failed (is the IAM service running at $IAM_URL?)"
  echo "Response: $LOGIN_RESP"
  exit 1
}

ACCESS_TOKEN=$(echo "$LOGIN_RESP" | jq -r '.accessToken // empty')
if [ -z "$ACCESS_TOKEN" ]; then
  fail "Admin login: no accessToken in response"
  echo "Response: $LOGIN_RESP"
  exit 1
fi
pass "Admin login succeeded"

# Step 2: Create M2M application
echo ""
echo "--- Step 2: Create M2M application ---"
APP_RESP=$(curl -sf --max-time 10 -X POST "$IAM_URL/v1/applications" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$APP_NAME\",
    \"type\": \"m2m\",
    \"grant_types\": [\"client_credentials\"],
    \"scopes\": [\"openid\"]
  }" 2>&1) || {
  fail "Create M2M application request failed"
  echo "Response: $APP_RESP"
  exit 1
}

APP_ID=$(echo "$APP_RESP" | jq -r '.application.id // empty')
CLIENT_ID=$(echo "$APP_RESP" | jq -r '.application.clientId // empty')
CLIENT_SECRET=$(echo "$APP_RESP" | jq -r '.clientSecret // empty')

if [ -z "$APP_ID" ] || [ -z "$CLIENT_ID" ] || [ -z "$CLIENT_SECRET" ]; then
  fail "Create M2M application: missing id/clientId/clientSecret in response"
  echo "Response: $APP_RESP"
  exit 1
fi
pass "M2M application created (id=$APP_ID, client_id=$CLIENT_ID)"

# Step 3: Get token via client_credentials grant
echo ""
echo "--- Step 3: Get token via client_credentials ---"
TOKEN_RESP=$(curl -sf --max-time 10 -X POST "$IAM_URL/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET&scope=openid" \
  2>&1) || {
  fail "client_credentials token request failed"
  echo "Response: $TOKEN_RESP"
  # Cleanup before exit
  curl -sf --max-time 5 -X DELETE "$IAM_URL/v1/applications/$APP_ID" \
    -H "Authorization: Bearer $ACCESS_TOKEN" &>/dev/null || true
  exit 1
}

# Step 4: Verify token response
echo ""
echo "--- Step 4: Verify token response ---"
M2M_ACCESS_TOKEN=$(echo "$TOKEN_RESP" | jq -r '.access_token // empty')
TOKEN_TYPE=$(echo "$TOKEN_RESP" | jq -r '.token_type // empty')
EXPIRES_IN=$(echo "$TOKEN_RESP" | jq -r '.expires_in // empty')

if [ -z "$M2M_ACCESS_TOKEN" ]; then
  fail "Token response missing access_token"
  echo "Response: $TOKEN_RESP"
else
  pass "access_token present in token response"
fi

if [ "$TOKEN_TYPE" = "Bearer" ]; then
  pass "token_type is Bearer"
else
  fail "token_type expected 'Bearer', got '$TOKEN_TYPE'"
fi

if [ -n "$EXPIRES_IN" ] && [ "$EXPIRES_IN" -gt 0 ] 2>/dev/null; then
  pass "expires_in=$EXPIRES_IN (positive)"
else
  fail "expires_in missing or invalid: '$EXPIRES_IN'"
fi

# Step 5: Cleanup — delete the M2M application
echo ""
echo "--- Step 5: Cleanup ---"
DEL_RESP=$(curl -sf --max-time 10 -X DELETE "$IAM_URL/v1/applications/$APP_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" 2>&1) || {
  fail "Delete M2M application failed (id=$APP_ID)"
  echo "Response: $DEL_RESP"
  # Don't exit; report results anyway
}

DEL_OK=$(echo "$DEL_RESP" | jq -r '.success // empty')
if [ "$DEL_OK" = "true" ]; then
  pass "M2M application deleted (id=$APP_ID)"
else
  fail "Delete did not return success=true: $DEL_RESP"
fi

# Summary
echo ""
echo "=== Results ==="
echo "PASS: $PASS"
echo "FAIL: $FAIL"
echo ""
if [ "$FAIL" -eq 0 ]; then
  echo "✓ All tests passed"
  exit 0
else
  echo "✗ $FAIL test(s) failed"
  exit 1
fi
