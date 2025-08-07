#!/bin/bash

# Multi-tenant Database Seeder Script
# Seeds data for all tenants or specific tenant

set -e

# Configuration
POSTGRES_HOST=${POSTGRES_HOST:-localhost}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_DB=${POSTGRES_DB:-zplus_saas}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-password}

MONGO_HOST=${MONGO_HOST:-localhost}
MONGO_PORT=${MONGO_PORT:-27017}

REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEEDERS_DIR="$SCRIPT_DIR/../database/seeders"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Check if PostgreSQL is available
check_postgres() {
    if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c '\q' > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Check if MongoDB is available
check_mongo() {
    if mongosh --host $MONGO_HOST:$MONGO_PORT --eval "db.runCommand('ping')" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Run PostgreSQL seeders
run_postgres_seeders() {
    local tenant_slug="$1"
    
    log_info "Running PostgreSQL seeders..."
    
    if [ -z "$tenant_slug" ]; then
        # Run all base seeders
        for seeder_file in $(ls "$SEEDERS_DIR"/*.sql | grep -v tenant_ | sort); do
            log_info "Running seeder: $(basename "$seeder_file")"
            PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -f "$seeder_file"
        done
        
        # Run tenant-specific seeders
        for seeder_file in $(ls "$SEEDERS_DIR"/tenant_*.sql 2>/dev/null | sort); do
            log_info "Running tenant seeder: $(basename "$seeder_file")"
            PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -f "$seeder_file"
        done
    else
        # Run specific tenant seeder
        tenant_seeder="$SEEDERS_DIR/tenant_${tenant_slug}_seeder.sql"
        if [ -f "$tenant_seeder" ]; then
            log_info "Running seeder for tenant: $tenant_slug"
            PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -f "$tenant_seeder"
        else
            log_error "Tenant seeder not found: $tenant_seeder"
            return 1
        fi
    fi
}

# Generate tenant seeder from template
generate_tenant_seeder() {
    local tenant_slug="$1"
    local tenant_id="$2"
    local tenant_name="$3"
    
    if [ -z "$tenant_slug" ] || [ -z "$tenant_id" ] || [ -z "$tenant_name" ]; then
        log_error "Usage: generate_tenant_seeder <tenant_slug> <tenant_id> <tenant_name>"
        return 1
    fi
    
    local template_file="$SEEDERS_DIR/007_tenant_acme_corp_seeder.sql"
    local output_file="$SEEDERS_DIR/tenant_${tenant_slug}_seeder.sql"
    
    if [ ! -f "$template_file" ]; then
        log_error "Template file not found: $template_file"
        return 1
    fi
    
    log_info "Generating seeder for tenant: $tenant_slug"
    
    # Replace template variables
    sed -e "s/acme_corp/$tenant_slug/g" \
        -e "s/demo_tenant_001/$tenant_id/g" \
        -e "s/ACME Corporation/$tenant_name/g" \
        -e "s/acme\.com/${tenant_slug}.com/g" \
        "$template_file" > "$output_file"
    
    log_info "Tenant seeder generated: $output_file"
}

# Seed MongoDB data for tenant
seed_mongo_tenant() {
    local tenant_slug="$1"
    local db_name="tenant_${tenant_slug}"
    
    if [ -z "$tenant_slug" ]; then
        log_error "Tenant slug is required"
        return 1
    fi
    
    log_info "Seeding MongoDB data for tenant: $tenant_slug"
    
    # Create and seed tenant MongoDB database
    mongosh --host $MONGO_HOST:$MONGO_PORT "$db_name" << EOF
// Sample Analytics Data
db.page_views.insertMany([
    {
        tenant_id: "$tenant_slug",
        user_id: "user_001",
        page: "/dashboard",
        timestamp: new Date(),
        session_id: "sess_001",
        ip_address: "192.168.1.100",
        user_agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        duration: 120,
        metadata: { referrer: "direct" }
    },
    {
        tenant_id: "$tenant_slug",
        user_id: "user_002",
        page: "/products",
        timestamp: new Date(Date.now() - 3600000),
        session_id: "sess_002",
        ip_address: "192.168.1.101",
        user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        duration: 180,
        metadata: { referrer: "google.com" }
    }
]);

// Sample User Events
db.user_events.insertMany([
    {
        tenant_id: "$tenant_slug",
        user_id: "user_001",
        event_type: "login",
        timestamp: new Date(),
        session_id: "sess_001",
        properties: { method: "email" },
        metadata: { ip: "192.168.1.100" }
    },
    {
        tenant_id: "$tenant_slug",
        user_id: "user_002",
        event_type: "view_product",
        timestamp: new Date(Date.now() - 1800000),
        session_id: "sess_002",
        properties: { product_id: "prod_001", category: "electronics" },
        metadata: { page: "/products/macbook-pro" }
    }
]);

// Sample Notifications
db.notifications.insertMany([
    {
        tenant_id: "$tenant_slug",
        user_id: "user_001",
        type: "info",
        title: "Welcome to ${tenant_slug}!",
        message: "Your account has been created successfully.",
        is_read: false,
        priority: "medium",
        created_at: new Date(),
        expires_at: new Date(Date.now() + 30 * 24 * 3600000) // 30 days
    },
    {
        tenant_id: "$tenant_slug",
        user_id: "user_002",
        type: "success",
        title: "Order Confirmed",
        message: "Your order #ORD-2025-0001 has been confirmed.",
        is_read: false,
        priority: "high",
        created_at: new Date(),
        expires_at: new Date(Date.now() + 7 * 24 * 3600000) // 7 days
    }
]);

// Sample Activity Logs
db.activity_logs.insertMany([
    {
        tenant_id: "$tenant_slug",
        user_id: "admin_001",
        action: "create",
        resource_type: "product",
        resource_id: "prod_001",
        timestamp: new Date(),
        ip_address: "192.168.1.100",
        user_agent: "Mozilla/5.0",
        changes: { name: "MacBook Pro 16\"", price: 2499.99 },
        metadata: { source: "admin_panel" }
    }
]);

print("MongoDB seed data created for tenant: $tenant_slug");
EOF
    
    log_info "MongoDB seeding completed for tenant: $tenant_slug"
}

# Seed Redis data for tenant
seed_redis_tenant() {
    local tenant_slug="$1"
    
    if [ -z "$tenant_slug" ]; then
        log_error "Tenant slug is required"
        return 1
    fi
    
    log_info "Seeding Redis data for tenant: $tenant_slug"
    
    # Set sample cache data
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET "${tenant_slug}:cache:products:featured" '["prod_001","prod_002","prod_003"]' EX 3600
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET "${tenant_slug}:cache:categories:all" '[{"id":"cat_1","name":"Electronics"},{"id":"cat_2","name":"Books"}]' EX 7200
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET "${tenant_slug}:metrics:realtime" '{"active_users":15,"orders_today":5,"revenue_today":3599.97}' EX 300
    
    # Set feature flags
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET "${tenant_slug}:features:new_dashboard" "true"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET "${tenant_slug}:features:advanced_search" "false"
    
    # Set online users
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SADD "${tenant_slug}:online:users" "user_001" "user_002"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT EXPIRE "${tenant_slug}:online:users" 300
    
    log_info "Redis seeding completed for tenant: $tenant_slug"
}

# Clear all seed data
clear_seed_data() {
    local tenant_slug="$1"
    
    log_warn "This will clear all seed data. Continue? (yes/no)"
    read -r confirm
    if [ "$confirm" != "yes" ]; then
        log_info "Operation cancelled"
        return 0
    fi
    
    if [ -z "$tenant_slug" ]; then
        # Clear all data
        log_warn "Clearing ALL seed data..."
        
        # Clear PostgreSQL
        if check_postgres; then
            PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB << EOF
TRUNCATE TABLE tenant_domains CASCADE;
TRUNCATE TABLE tenants CASCADE;
EOF
        fi
        
        # Clear MongoDB (all tenant databases)
        if check_mongo; then
            mongosh --host $MONGO_HOST:$MONGO_PORT --eval "
                db.adminCommand('listDatabases').databases.forEach(function(database) {
                    if (database.name.startsWith('tenant_')) {
                        db.getSiblingDB(database.name).dropDatabase();
                        print('Dropped database: ' + database.name);
                    }
                });
            "
        fi
        
        # Clear Redis (all tenant keys)
        redis-cli -h $REDIS_HOST -p $REDIS_PORT EVAL "
            local cursor = '0'
            repeat
                local scan_result = redis.call('SCAN', cursor, 'MATCH', '*:*')
                cursor = scan_result[1]
                local keys = scan_result[2]
                if #keys > 0 then
                    redis.call('DEL', unpack(keys))
                end
            until cursor == '0'
            return 'OK'
        " 0
        
    else
        # Clear specific tenant data
        log_warn "Clearing seed data for tenant: $tenant_slug"
        
        # Clear PostgreSQL tenant schema
        if check_postgres; then
            PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT drop_tenant_schema('$tenant_slug');"
        fi
        
        # Clear MongoDB tenant database
        if check_mongo; then
            mongosh --host $MONGO_HOST:$MONGO_PORT "tenant_${tenant_slug}" --eval "db.dropDatabase();"
        fi
        
        # Clear Redis tenant keys
        redis-cli -h $REDIS_HOST -p $REDIS_PORT EVAL "
            local keys = redis.call('keys', ARGV[1])
            if #keys > 0 then
                return redis.call('del', unpack(keys))
            end
            return 0
        " 0 "${tenant_slug}:*"
    fi
    
    log_info "Seed data cleared successfully"
}

# Show seeding status
show_status() {
    log_info "Database Seeding Status:"
    echo "========================"
    
    if check_postgres; then
        echo "PostgreSQL:"
        tenant_count=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL;" | tr -d ' ')
        echo "  - Active tenants: $tenant_count"
        
        if [ "$tenant_count" -gt 0 ]; then
            PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "
                SELECT slug as tenant, 
                       (SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = 'tenant_' || slug) as has_schema
                FROM tenants 
                WHERE deleted_at IS NULL 
                ORDER BY created_at;
            "
        fi
    fi
    
    if check_mongo; then
        echo ""
        echo "MongoDB:"
        mongo_tenant_count=$(mongosh --host $MONGO_HOST:$MONGO_PORT --quiet --eval "
            db.adminCommand('listDatabases').databases.filter(db => db.name.startsWith('tenant_')).length
        ")
        echo "  - Tenant databases: $mongo_tenant_count"
    fi
    
    echo ""
    echo "Redis:"
    redis_tenant_count=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT EVAL "
        local tenants = {}
        local cursor = '0'
        repeat
            local scan_result = redis.call('SCAN', cursor, 'MATCH', '*:*')
            cursor = scan_result[1]
            local keys = scan_result[2]
            for i = 1, #keys do
                local tenant = string.match(keys[i], '^([^:]+):')
                if tenant then
                    tenants[tenant] = true
                end
            end
        until cursor == '0'
        
        local count = 0
        for k, v in pairs(tenants) do
            count = count + 1
        end
        return count
    " 0)
    echo "  - Tenants with Redis data: $redis_tenant_count"
}

# Main function
main() {
    case "${1:-help}" in
        "all")
            if ! check_postgres; then
                log_error "PostgreSQL connection failed"
                exit 1
            fi
            run_postgres_seeders
            
            # Seed sample tenants in MongoDB and Redis
            for tenant in "acme_corp" "startup_inc" "enterprise_sol"; do
                if check_mongo; then
                    seed_mongo_tenant "$tenant"
                fi
                seed_redis_tenant "$tenant"
            done
            ;;
        "postgres")
            if ! check_postgres; then
                log_error "PostgreSQL connection failed"
                exit 1
            fi
            run_postgres_seeders "$2"
            ;;
        "mongo")
            if [ -z "$2" ]; then
                log_error "Usage: $0 mongo <tenant_slug>"
                exit 1
            fi
            if ! check_mongo; then
                log_error "MongoDB connection failed"
                exit 1
            fi
            seed_mongo_tenant "$2"
            ;;
        "redis")
            if [ -z "$2" ]; then
                log_error "Usage: $0 redis <tenant_slug>"
                exit 1
            fi
            seed_redis_tenant "$2"
            ;;
        "tenant")
            if [ -z "$2" ]; then
                log_error "Usage: $0 tenant <tenant_slug>"
                exit 1
            fi
            tenant_slug="$2"
            
            if check_postgres; then
                run_postgres_seeders "$tenant_slug"
            fi
            if check_mongo; then
                seed_mongo_tenant "$tenant_slug"
            fi
            seed_redis_tenant "$tenant_slug"
            ;;
        "generate")
            if [ $# -lt 4 ]; then
                log_error "Usage: $0 generate <tenant_slug> <tenant_id> <tenant_name>"
                exit 1
            fi
            generate_tenant_seeder "$2" "$3" "$4"
            ;;
        "clear")
            clear_seed_data "$2"
            ;;
        "status")
            show_status
            ;;
        "help"|*)
            echo "Multi-tenant Database Seeder"
            echo "============================"
            echo ""
            echo "Usage: $0 <command> [options]"
            echo ""
            echo "Commands:"
            echo "  all                                 - Seed all databases with sample data"
            echo "  postgres [tenant_slug]              - Seed PostgreSQL (all or specific tenant)"
            echo "  mongo <tenant_slug>                 - Seed MongoDB for specific tenant"
            echo "  redis <tenant_slug>                 - Seed Redis for specific tenant"
            echo "  tenant <tenant_slug>                - Seed all databases for specific tenant"
            echo "  generate <slug> <id> <name>         - Generate tenant seeder from template"
            echo "  clear [tenant_slug]                 - Clear seed data (all or specific tenant)"
            echo "  status                              - Show seeding status"
            echo ""
            echo "Examples:"
            echo "  $0 all"
            echo "  $0 tenant acme_corp"
            echo "  $0 generate new_tenant tenant_004 'New Company Inc'"
            echo "  $0 clear acme_corp"
            ;;
    esac
}

main "$@"
