#!/bin/bash

# RBAC Migration Script
# This script runs the RBAC migration and initializes default roles and permissions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Database configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-zplus}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}

echo -e "${GREEN}🚀 Starting RBAC Migration...${NC}"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}❌ psql is not installed. Please install PostgreSQL client.${NC}"
    exit 1
fi

# Check database connection
echo -e "${YELLOW}📡 Testing database connection...${NC}"
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}❌ Cannot connect to database. Please check your configuration.${NC}"
    echo "Database: $DB_NAME"
    echo "Host: $DB_HOST:$DB_PORT"
    echo "User: $DB_USER"
    exit 1
fi

echo -e "${GREEN}✅ Database connection successful${NC}"

# Run the RBAC migration
echo -e "${YELLOW}📦 Running RBAC migration...${NC}"
MIGRATION_FILE="$(dirname "$0")/migrations/007_create_rbac_tables.sql"

if [ ! -f "$MIGRATION_FILE" ]; then
    echo -e "${RED}❌ Migration file not found: $MIGRATION_FILE${NC}"
    exit 1
fi

if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$MIGRATION_FILE"; then
    echo -e "${GREEN}✅ RBAC migration completed successfully${NC}"
else
    echo -e "${RED}❌ RBAC migration failed${NC}"
    exit 1
fi

# Verify the migration
echo -e "${YELLOW}🔍 Verifying migration...${NC}"

# Check if tables were created
TABLES=("permissions" "roles" "role_permissions" "user_roles" "casbin_rule")
for table in "${TABLES[@]}"; do
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1 FROM $table LIMIT 1;" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Table $table created successfully${NC}"
    else
        echo -e "${RED}❌ Table $table not found or empty${NC}"
        exit 1
    fi
done

# Check if default permissions were inserted
PERM_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM permissions;")
if [ "$PERM_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Default permissions inserted ($PERM_COUNT permissions)${NC}"
else
    echo -e "${RED}❌ No permissions found${NC}"
    exit 1
fi

# Check if default roles were inserted
ROLE_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM roles WHERE is_system = true;")
if [ "$ROLE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Default system roles inserted ($ROLE_COUNT roles)${NC}"
else
    echo -e "${RED}❌ No system roles found${NC}"
    exit 1
fi

# List created permissions
echo -e "${YELLOW}📋 Created permissions:${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    name,
    resource,
    action,
    description
FROM permissions 
ORDER BY resource, action;"

# List created roles
echo -e "${YELLOW}📋 Created system roles:${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    r.name,
    r.description,
    COUNT(rp.permission_id) as permission_count
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
WHERE r.is_system = true
GROUP BY r.id, r.name, r.description
ORDER BY r.name;"

echo -e "${GREEN}🎉 RBAC setup completed successfully!${NC}"
echo -e "${YELLOW}📚 Next steps:${NC}"
echo "1. Update your application to use the new RBAC system"
echo "2. Create tenant-specific roles for new tenants"
echo "3. Assign roles to users"
echo "4. Test the permission system"

echo -e "${GREEN}🔒 RBAC System Features:${NC}"
echo "• System-level roles for admin access"
echo "• Tenant-specific roles for multi-tenant isolation"
echo "• Fine-grained permissions for resource access"
echo "• Casbin integration for policy enforcement"
echo "• GraphQL API for role and permission management"
