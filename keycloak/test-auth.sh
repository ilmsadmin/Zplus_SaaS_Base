#!/bin/bash

# Zplus SaaS Base - Authentication Test Script
# This script tests the Keycloak authentication integration

set -e

KEYCLOAK_URL="http://localhost:8081"
REALM_NAME="zplus"
CLIENT_ID="zplus-backend"
CLIENT_SECRET="zplus-backend-secret-2024"

echo "🧪 Testing Keycloak Authentication Integration"
echo "=============================================="

# Test 1: Check if Keycloak is running
echo ""
echo "1️⃣  Testing Keycloak Health..."
if curl -s -f "$KEYCLOAK_URL/health/ready" > /dev/null; then
    echo "✅ Keycloak is healthy and ready"
else
    echo "❌ Keycloak is not ready"
    echo "   Make sure to run: make dev-up"
    exit 1
fi

# Test 2: Check realm configuration
echo ""
echo "2️⃣  Testing Realm Configuration..."
REALM_CONFIG=$(curl -s "$KEYCLOAK_URL/realms/$REALM_NAME/.well-known/openid_configuration")
if echo "$REALM_CONFIG" | jq -e '.issuer' > /dev/null 2>&1; then
    ISSUER=$(echo "$REALM_CONFIG" | jq -r '.issuer')
    echo "✅ Realm '$REALM_NAME' is configured correctly"
    echo "   Issuer: $ISSUER"
else
    echo "❌ Realm '$REALM_NAME' is not configured"
    echo "   Run: make keycloak-setup"
    exit 1
fi

# Test 3: Test user authentication
echo ""
echo "3️⃣  Testing User Authentication..."

# Test System Admin
echo "   Testing system admin login..."
ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=system.admin" \
    -d "password=Admin123!" \
    -d "grant_type=password" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET" | jq -r '.access_token // "null"')

if [ "$ADMIN_TOKEN" != "null" ] && [ "$ADMIN_TOKEN" != "" ]; then
    echo "✅ System admin authentication successful"
    
    # Decode and display token claims
    PAYLOAD=$(echo "$ADMIN_TOKEN" | cut -d. -f2)
    # Add padding if needed
    while [ $((${#PAYLOAD} % 4)) -ne 0 ]; do
        PAYLOAD="${PAYLOAD}="
    done
    DECODED=$(echo "$PAYLOAD" | base64 -d 2>/dev/null | jq . 2>/dev/null || echo "Could not decode token")
    if [ "$DECODED" != "Could not decode token" ]; then
        echo "   Token claims:"
        echo "$DECODED" | jq -r '"     Subject: " + .sub'
        echo "$DECODED" | jq -r '"     Email: " + .email'
        echo "$DECODED" | jq -r '"     Roles: " + (.realm_access.roles | join(", "))'
        echo "$DECODED" | jq -r '"     Tenant ID: " + (.tenant_id // "N/A")'
    fi
else
    echo "❌ System admin authentication failed"
fi

# Test Tenant Admin
echo ""
echo "   Testing tenant admin login..."
TENANT_ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=tenant.admin" \
    -d "password=TenantAdmin123!" \
    -d "grant_type=password" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET" | jq -r '.access_token // "null"')

if [ "$TENANT_ADMIN_TOKEN" != "null" ] && [ "$TENANT_ADMIN_TOKEN" != "" ]; then
    echo "✅ Tenant admin authentication successful"
else
    echo "❌ Tenant admin authentication failed"
fi

# Test Tenant User
echo ""
echo "   Testing tenant user login..."
USER_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=john.doe" \
    -d "password=User123!" \
    -d "grant_type=password" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET" | jq -r '.access_token // "null"')

if [ "$USER_TOKEN" != "null" ] && [ "$USER_TOKEN" != "" ]; then
    echo "✅ Tenant user authentication successful"
else
    echo "❌ Tenant user authentication failed"
fi

# Test 4: Check JWK endpoint
echo ""
echo "4️⃣  Testing JWK Endpoint..."
JWK_RESPONSE=$(curl -s "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/certs")
if echo "$JWK_RESPONSE" | jq -e '.keys' > /dev/null 2>&1; then
    KEY_COUNT=$(echo "$JWK_RESPONSE" | jq '.keys | length')
    echo "✅ JWK endpoint is working"
    echo "   Available keys: $KEY_COUNT"
else
    echo "❌ JWK endpoint is not working"
fi

# Test 5: Test invalid credentials
echo ""
echo "5️⃣  Testing Invalid Credentials..."
INVALID_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=invalid.user" \
    -d "password=wrongpassword" \
    -d "grant_type=password" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET" | jq -r '.access_token // "null"')

if [ "$INVALID_TOKEN" = "null" ]; then
    echo "✅ Invalid credentials properly rejected"
else
    echo "❌ Invalid credentials not properly rejected"
fi

# Summary
echo ""
echo "📊 Test Summary"
echo "==============="
echo "Keycloak URL: $KEYCLOAK_URL"
echo "Realm: $REALM_NAME"
echo "Client ID: $CLIENT_ID"
echo ""

if [ "$ADMIN_TOKEN" != "null" ] && [ "$TENANT_ADMIN_TOKEN" != "null" ] && [ "$USER_TOKEN" != "null" ]; then
    echo "🎉 All authentication tests passed!"
    echo ""
    echo "🔗 Useful URLs:"
    echo "   Admin Console: $KEYCLOAK_URL"
    echo "   Realm Config: $KEYCLOAK_URL/realms/$REALM_NAME/.well-known/openid_configuration"
    echo "   Token Endpoint: $KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token"
    echo "   JWK Endpoint: $KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/certs"
    echo ""
    echo "📝 Sample curl command to get token:"
    echo "curl -X POST '$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token' \\"
    echo "  -H 'Content-Type: application/x-www-form-urlencoded' \\"
    echo "  -d 'username=john.doe' \\"
    echo "  -d 'password=User123!' \\"
    echo "  -d 'grant_type=password' \\"
    echo "  -d 'client_id=$CLIENT_ID' \\"
    echo "  -d 'client_secret=$CLIENT_SECRET'"
else
    echo "❌ Some authentication tests failed"
    echo "   Please check the setup and try again"
    exit 1
fi
