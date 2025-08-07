#!/bin/bash

# Zplus SaaS Base - Keycloak Setup Script
# This script configures Keycloak for multi-tenant authentication

set -e

KEYCLOAK_URL="http://localhost:8081"
ADMIN_USER="admin"
ADMIN_PASSWORD="admin123"
REALM_NAME="zplus"

echo "🔐 Setting up Keycloak for Zplus SaaS..."

# Wait for Keycloak to be ready
echo "⏳ Waiting for Keycloak to be ready..."
until curl -s -f "$KEYCLOAK_URL/health/ready" > /dev/null 2>&1; do
    echo "   Waiting for Keycloak..."
    sleep 5
done

echo "✅ Keycloak is ready!"

# Get admin token
echo "🔑 Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=$ADMIN_USER" \
    -d "password=$ADMIN_PASSWORD" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" | jq -r '.access_token')

if [ "$ADMIN_TOKEN" = "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Failed to get admin token"
    exit 1
fi

echo "✅ Admin token obtained"

# Check if realm exists
echo "🔍 Checking if realm '$REALM_NAME' exists..."
REALM_EXISTS=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
    "$KEYCLOAK_URL/admin/realms/$REALM_NAME" | jq -r '.realm // "null"')

if [ "$REALM_EXISTS" = "null" ]; then
    echo "📥 Importing realm configuration..."
    curl -s -X POST "$KEYCLOAK_URL/admin/realms" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d @/opt/keycloak/realm-config/zplus-realm.json
    
    if [ $? -eq 0 ]; then
        echo "✅ Realm '$REALM_NAME' imported successfully"
    else
        echo "❌ Failed to import realm"
        exit 1
    fi
else
    echo "ℹ️  Realm '$REALM_NAME' already exists"
fi

# Create default users
echo "👤 Creating default users..."

# System Admin User
echo "   Creating system admin user..."
SYSTEM_ADMIN_JSON='{
    "username": "system.admin",
    "email": "admin@zplus.io",
    "firstName": "System",
    "lastName": "Administrator",
    "enabled": true,
    "emailVerified": true,
    "credentials": [{
        "type": "password",
        "value": "Admin123!",
        "temporary": false
    }],
    "attributes": {
        "tenant_id": ["system"],
        "tenant_domain": ["admin.zplus.io"],
        "tenant_permissions": ["{\"system:manage\": true, \"tenant:create\": true, \"tenant:manage\": true, \"user:manage_all\": true}"]
    },
    "groups": ["/system-admins"],
    "realmRoles": ["system_admin"]
}'

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$SYSTEM_ADMIN_JSON" > /dev/null 2>&1

# Demo Tenant Admin User
echo "   Creating demo tenant admin user..."
TENANT_ADMIN_JSON='{
    "username": "tenant.admin",
    "email": "admin@acme.example.com",
    "firstName": "Tenant",
    "lastName": "Administrator",
    "enabled": true,
    "emailVerified": true,
    "credentials": [{
        "type": "password",
        "value": "TenantAdmin123!",
        "temporary": false
    }],
    "attributes": {
        "tenant_id": ["acme_corp"],
        "tenant_domain": ["acme.zplus.io"],
        "tenant_permissions": ["{\"tenant:manage_own\": true, \"user:manage_tenant\": true, \"domain:manage\": true}"]
    },
    "groups": ["/tenant-admins"],
    "realmRoles": ["tenant_admin"]
}'

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$TENANT_ADMIN_JSON" > /dev/null 2>&1

# Demo Tenant User
echo "   Creating demo tenant user..."
TENANT_USER_JSON='{
    "username": "john.doe",
    "email": "john.doe@acme.example.com",
    "firstName": "John",
    "lastName": "Doe",
    "enabled": true,
    "emailVerified": true,
    "credentials": [{
        "type": "password",
        "value": "User123!",
        "temporary": false
    }],
    "attributes": {
        "tenant_id": ["acme_corp"],
        "tenant_domain": ["acme.zplus.io"],
        "tenant_permissions": ["{\"tenant:access\": true, \"profile:manage_own\": true}"]
    },
    "groups": ["/tenant-users"],
    "realmRoles": ["tenant_user"]
}'

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$TENANT_USER_JSON" > /dev/null 2>&1

echo "✅ Default users created"

# Display login information
echo ""
echo "🎉 Keycloak setup completed successfully!"
echo ""
echo "📋 Login Information:"
echo "   Keycloak Admin Console: $KEYCLOAK_URL"
echo "   Admin Username: $ADMIN_USER"
echo "   Admin Password: $ADMIN_PASSWORD"
echo ""
echo "🔐 Test Users:"
echo "   System Admin: system.admin / Admin123!"
echo "   Tenant Admin: tenant.admin / TenantAdmin123!"
echo "   Tenant User:  john.doe / User123!"
echo ""
echo "🌐 Realm: $REALM_NAME"
echo "   Realm URL: $KEYCLOAK_URL/realms/$REALM_NAME"
echo ""
echo "🔑 Client Information:"
echo "   Backend Client ID: zplus-backend"
echo "   Admin Frontend Client ID: zplus-admin-frontend"
echo "   Tenant Frontend Client ID: zplus-tenant-frontend"
echo ""
