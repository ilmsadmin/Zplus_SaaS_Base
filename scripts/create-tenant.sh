#!/bin/bash

# Zplus SaaS Base - Create Tenant Script
# Description: Creates a new tenant with database schemas and configurations
# Usage: ./create-tenant.sh <tenant_id>

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load environment variables
if [ -f "$PROJECT_ROOT/.env" ]; then
    source "$PROJECT_ROOT/.env"
fi

# Default values
POSTGRES_HOST=${POSTGRES_HOST:-localhost}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres123}
POSTGRES_DB=${POSTGRES_DB:-zplus}

MONGODB_HOST=${MONGODB_HOST:-localhost}
MONGODB_PORT=${MONGODB_PORT:-27017}
MONGODB_USERNAME=${MONGODB_USERNAME:-mongo}
MONGODB_PASSWORD=${MONGODB_PASSWORD:-mongo123}

REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-redis123}

# Functions
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

show_usage() {
    echo "Usage: $0 <tenant_id>"
    echo ""
    echo "Arguments:"
    echo "  tenant_id    Unique identifier for the tenant (alphanumeric, lowercase)"
    echo ""
    echo "Examples:"
    echo "  $0 acme_corp"
    echo "  $0 company123"
    echo ""
    echo "Environment variables (optional):"
    echo "  POSTGRES_HOST     PostgreSQL host (default: localhost)"
    echo "  POSTGRES_PORT     PostgreSQL port (default: 5432)"
    echo "  POSTGRES_USER     PostgreSQL user (default: postgres)"
    echo "  POSTGRES_PASSWORD PostgreSQL password"
    echo "  MONGODB_HOST      MongoDB host (default: localhost)"
    echo "  MONGODB_PORT      MongoDB port (default: 27017)"
    echo "  REDIS_HOST        Redis host (default: localhost)"
    echo "  REDIS_PORT        Redis port (default: 6379)"
}

validate_tenant_id() {
    local tenant_id=$1
    
    if [[ ! "$tenant_id" =~ ^[a-z0-9_]+$ ]]; then
        log_error "Tenant ID must contain only lowercase letters, numbers, and underscores"
        return 1
    fi
    
    if [[ ${#tenant_id} -lt 3 || ${#tenant_id} -gt 63 ]]; then
        log_error "Tenant ID must be between 3 and 63 characters"
        return 1
    fi
    
    return 0
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    command -v psql >/dev/null 2>&1 || {
        log_error "psql is required but not installed. Please install PostgreSQL client."
        exit 1
    }
    
    command -v mongosh >/dev/null 2>&1 || command -v mongo >/dev/null 2>&1 || {
        log_error "mongosh or mongo is required but not installed. Please install MongoDB client."
        exit 1
    }
    
    command -v redis-cli >/dev/null 2>&1 || {
        log_warning "redis-cli not found. Redis operations will be skipped."
    }
    
    log_success "Dependencies checked"
}

test_connections() {
    log_info "Testing database connections..."
    
    # Test PostgreSQL connection
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q' >/dev/null 2>&1 || {
        log_error "Cannot connect to PostgreSQL database"
        log_error "Host: $POSTGRES_HOST:$POSTGRES_PORT, User: $POSTGRES_USER, Database: $POSTGRES_DB"
        exit 1
    }
    
    # Test MongoDB connection
    if command -v mongosh >/dev/null 2>&1; then
        mongosh --host "$MONGODB_HOST:$MONGODB_PORT" --username "$MONGODB_USERNAME" --password "$MONGODB_PASSWORD" --eval "db.adminCommand('ping')" >/dev/null 2>&1 || {
            log_error "Cannot connect to MongoDB database"
            exit 1
        }
    else
        mongo --host "$MONGODB_HOST:$MONGODB_PORT" --username "$MONGODB_USERNAME" --password "$MONGODB_PASSWORD" --eval "db.adminCommand('ping')" >/dev/null 2>&1 || {
            log_error "Cannot connect to MongoDB database"
            exit 1
        }
    fi
    
    # Test Redis connection (optional)
    if command -v redis-cli >/dev/null 2>&1; then
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" ping >/dev/null 2>&1 || {
            log_warning "Cannot connect to Redis. Redis operations will be skipped."
        }
    fi
    
    log_success "Database connections tested"
}

create_postgres_schema() {
    local tenant_id=$1
    
    log_info "Creating PostgreSQL schema for tenant: $tenant_id"
    
    # Check if schema already exists
    local schema_exists=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '$tenant_id';")
    
    if [[ "$schema_exists" -gt 0 ]]; then
        log_warning "PostgreSQL schema '$tenant_id' already exists"
        return 0
    fi
    
    # Create schema
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << EOF
-- Create schema for tenant
CREATE SCHEMA IF NOT EXISTS $tenant_id;

-- Create tenant domains table if not exists
CREATE TABLE IF NOT EXISTS public.tenant_domains (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    domain VARCHAR(255) NOT NULL UNIQUE,
    is_custom BOOLEAN DEFAULT FALSE,
    verified BOOLEAN DEFAULT FALSE,
    ssl_enabled BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default subdomain for tenant
INSERT INTO public.tenant_domains (tenant_id, domain, is_custom, verified, ssl_enabled)
VALUES ('$tenant_id', '$tenant_id.zplus.io', FALSE, TRUE, TRUE);

-- Create tenant configuration table
CREATE TABLE IF NOT EXISTS $tenant_id.tenant_config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create users table for tenant
CREATE TABLE IF NOT EXISTS $tenant_id.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    active BOOLEAN DEFAULT TRUE,
    role VARCHAR(50) DEFAULT 'user',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create roles table for tenant
CREATE TABLE IF NOT EXISTS $tenant_id.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create user_roles junction table
CREATE TABLE IF NOT EXISTS $tenant_id.user_roles (
    user_id UUID NOT NULL REFERENCES $tenant_id.users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES $tenant_id.roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Create files table for tenant
CREATE TABLE IF NOT EXISTS $tenant_id.files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100),
    size BIGINT,
    uploaded_by UUID REFERENCES $tenant_id.users(id),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_${tenant_id}_users_email ON $tenant_id.users(email);
CREATE INDEX IF NOT EXISTS idx_${tenant_id}_users_active ON $tenant_id.users(active);
CREATE INDEX IF NOT EXISTS idx_${tenant_id}_users_role ON $tenant_id.users(role);
CREATE INDEX IF NOT EXISTS idx_${tenant_id}_files_uploaded_by ON $tenant_id.files(uploaded_by);

-- Insert default configuration
INSERT INTO $tenant_id.tenant_config (key, value) VALUES
    ('tenant_id', '$tenant_id'),
    ('tenant_name', '$tenant_id'),
    ('subdomain', '$tenant_id.zplus.io'),
    ('status', 'active'),
    ('custom_domain_enabled', 'true'),
    ('created_at', CURRENT_TIMESTAMP::text);

-- Insert default roles
INSERT INTO $tenant_id.roles (name, description, permissions) VALUES
    ('tenant_admin', 'Tenant Administrator with full access', '["*"]'),
    ('tenant_manager', 'Tenant Manager with limited admin access', '["user:read", "user:write", "file:read", "file:write"]'),
    ('user', 'Standard user role', '["file:read", "file:write", "profile:read", "profile:write"]'),
    ('viewer', 'Read-only access', '["file:read", "profile:read"]');

-- Grant permissions
GRANT USAGE ON SCHEMA $tenant_id TO $POSTGRES_USER;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA $tenant_id TO $POSTGRES_USER;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA $tenant_id TO $POSTGRES_USER;

COMMIT;
EOF
    
    if [ $? -eq 0 ]; then
        log_success "PostgreSQL schema created for tenant: $tenant_id"
    else
        log_error "Failed to create PostgreSQL schema for tenant: $tenant_id"
        exit 1
    fi
}

create_mongodb_database() {
    local tenant_id=$1
    local db_name="${tenant_id}_metadata"
    
    log_info "Creating MongoDB database for tenant: $tenant_id"
    
    # Create MongoDB database and collections
    if command -v mongosh >/dev/null 2>&1; then
        mongosh --host "$MONGODB_HOST:$MONGODB_PORT" --username "$MONGODB_USERNAME" --password "$MONGODB_PASSWORD" << EOF
use $db_name;

// Create collections with validation
db.createCollection("tenant_config", {
    validator: {
        \$jsonSchema: {
            bsonType: "object",
            required: ["key", "value"],
            properties: {
                key: { bsonType: "string" },
                value: { bsonType: ["string", "object", "array"] },
                createdAt: { bsonType: "date" },
                updatedAt: { bsonType: "date" }
            }
        }
    }
});

db.createCollection("user_preferences", {
    validator: {
        \$jsonSchema: {
            bsonType: "object",
            required: ["userId"],
            properties: {
                userId: { bsonType: "string" },
                preferences: { bsonType: "object" },
                createdAt: { bsonType: "date" },
                updatedAt: { bsonType: "date" }
            }
        }
    }
});

db.createCollection("audit_logs", {
    validator: {
        \$jsonSchema: {
            bsonType: "object",
            required: ["action", "userId", "timestamp"],
            properties: {
                action: { bsonType: "string" },
                userId: { bsonType: "string" },
                resource: { bsonType: "string" },
                metadata: { bsonType: "object" },
                timestamp: { bsonType: "date" }
            }
        }
    }
});

// Insert default configuration
db.tenant_config.insertMany([
    {
        key: "tenant_id",
        value: "$tenant_id",
        createdAt: new Date(),
        updatedAt: new Date()
    },
    {
        key: "database_name",
        value: "$db_name",
        createdAt: new Date(),
        updatedAt: new Date()
    },
    {
        key: "features",
        value: {
            file_upload: true,
            user_management: true,
            analytics: true
        },
        createdAt: new Date(),
        updatedAt: new Date()
    }
]);

// Create indexes
db.user_preferences.createIndex({ "userId": 1 }, { unique: true });
db.audit_logs.createIndex({ "userId": 1, "timestamp": -1 });
db.audit_logs.createIndex({ "action": 1, "timestamp": -1 });

quit();
EOF
    else
        mongo --host "$MONGODB_HOST:$MONGODB_PORT" --username "$MONGODB_USERNAME" --password "$MONGODB_PASSWORD" << EOF
use $db_name;

// Create collections and insert data
db.tenant_config.insertMany([
    {
        key: "tenant_id",
        value: "$tenant_id",
        createdAt: new Date(),
        updatedAt: new Date()
    },
    {
        key: "database_name", 
        value: "$db_name",
        createdAt: new Date(),
        updatedAt: new Date()
    }
]);

db.user_preferences.createIndex({ "userId": 1 }, { unique: true });
db.audit_logs.createIndex({ "userId": 1, "timestamp": -1 });

quit();
EOF
    fi
    
    if [ $? -eq 0 ]; then
        log_success "MongoDB database created for tenant: $tenant_id"
    else
        log_error "Failed to create MongoDB database for tenant: $tenant_id"
        exit 1
    fi
}

setup_redis_keys() {
    local tenant_id=$1
    
    if ! command -v redis-cli >/dev/null 2>&1; then
        log_warning "redis-cli not available, skipping Redis setup"
        return 0
    fi
    
    log_info "Setting up Redis keys for tenant: $tenant_id"
    
    # Set tenant configuration in Redis
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" << EOF
SET ${tenant_id}:config:tenant_id "$tenant_id"
SET ${tenant_id}:config:status "active"
SET ${tenant_id}:config:created_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
EXPIRE ${tenant_id}:config:tenant_id 86400
EXPIRE ${tenant_id}:config:status 86400
EXPIRE ${tenant_id}:config:created_at 86400
EOF
    
    if [ $? -eq 0 ]; then
        log_success "Redis keys set up for tenant: $tenant_id"
    else
        log_warning "Failed to set up Redis keys for tenant: $tenant_id"
    fi
}

create_tenant_directory() {
    local tenant_id=$1
    local tenant_dir="$PROJECT_ROOT/tenants/$tenant_id"
    
    log_info "Creating tenant directory: $tenant_dir"
    
    mkdir -p "$tenant_dir"/{config,uploads,logs,backups}
    
    # Create tenant-specific configuration
    cat > "$tenant_dir/config/tenant.json" << EOF
{
  "tenant_id": "$tenant_id",
  "name": "$tenant_id",
  "status": "active",
  "database": {
    "postgres_schema": "$tenant_id",
    "mongodb_database": "${tenant_id}_metadata"
  },
  "features": {
    "file_upload": true,
    "user_management": true,
    "analytics": true,
    "custom_branding": false
  },
  "limits": {
    "max_users": 1000,
    "max_storage_mb": 10240,
    "max_api_calls_per_hour": 10000
  },
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "updated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    # Create README for tenant
    cat > "$tenant_dir/README.md" << EOF
# Tenant: $tenant_id

This directory contains tenant-specific configurations and data for \`$tenant_id\`.

## Directory Structure

- \`config/\` - Tenant configuration files
- \`uploads/\` - Uploaded files (if using local storage)
- \`logs/\` - Tenant-specific logs
- \`backups/\` - Tenant backup files

## Database Information

- **PostgreSQL Schema**: \`$tenant_id\`
- **MongoDB Database**: \`${tenant_id}_metadata\`
- **Redis Key Prefix**: \`${tenant_id}:\`

## Created

- **Date**: $(date -u +%Y-%m-%d)
- **Time**: $(date -u +%H:%M:%S) UTC

## Configuration

See \`config/tenant.json\` for detailed tenant configuration.
EOF
    
    log_success "Tenant directory created: $tenant_dir"
}

# Main execution
main() {
    echo -e "${BLUE}Zplus SaaS Base - Create Tenant Script${NC}"
    echo "========================================"
    
    # Check arguments
    if [ $# -ne 1 ]; then
        log_error "Missing tenant ID argument"
        echo ""
        show_usage
        exit 1
    fi
    
    local tenant_id=$1
    
    # Validate tenant ID
    if ! validate_tenant_id "$tenant_id"; then
        exit 1
    fi
    
    log_info "Creating tenant: $tenant_id"
    
    # Check dependencies and connections
    check_dependencies
    test_connections
    
    # Create tenant resources
    create_postgres_schema "$tenant_id"
    create_mongodb_database "$tenant_id"
    setup_redis_keys "$tenant_id"
    create_tenant_directory "$tenant_id"
    
    echo ""
    log_success "Tenant '$tenant_id' created successfully!"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Update your application configuration to recognize the new tenant"
    echo "2. Create initial admin user for the tenant"
    echo "3. Configure tenant-specific settings in the admin panel"
    echo "4. Test the tenant setup by accessing: https://$tenant_id.zplus.io"
    echo "5. (Optional) Add custom domain via admin panel"
    echo ""
    echo -e "${BLUE}Tenant Information:${NC}"
    echo "- Tenant ID: $tenant_id"
    echo "- Subdomain: $tenant_id.zplus.io"
    echo "- PostgreSQL Schema: $tenant_id"
    echo "- MongoDB Database: ${tenant_id}_metadata"
    echo "- Redis Key Prefix: ${tenant_id}:"
    echo "- Config Directory: $PROJECT_ROOT/tenants/$tenant_id"
    echo ""
    echo -e "${BLUE}Access URLs:${NC}"
    echo "- User Login: https://$tenant_id.zplus.io/login"
    echo "- Tenant Admin: https://$tenant_id.zplus.io/admin/login"
    echo "- User Dashboard: https://$tenant_id.zplus.io/dashboard"
    echo "- Admin Dashboard: https://$tenant_id.zplus.io/admin/dashboard"
}

# Run main function with all arguments
main "$@"
