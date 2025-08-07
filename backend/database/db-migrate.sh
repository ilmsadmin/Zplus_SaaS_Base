#!/bin/bash

#!/bin/bash

# Database Migration Script for Zplus SaaS Base
# This script handles PostgreSQL migrations, MongoDB setup, and tenant management

set -e

# Configuration
POSTGRES_HOST=${POSTGRES_HOST:-localhost}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_DB=${POSTGRES_DB:-zplus_saas}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-password}

MONGO_HOST=${MONGO_HOST:-localhost}
MONGO_PORT=${MONGO_PORT:-27017}
MONGO_DB=${MONGO_DB:-zplus_saas}

REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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
    log_info "Checking PostgreSQL connection..."
    if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c '\q' > /dev/null 2>&1; then
        log_info "PostgreSQL connection successful"
        return 0
    else
        log_error "Failed to connect to PostgreSQL"
        return 1
    fi
}

# Run PostgreSQL migrations
run_postgres_migrations() {
    log_info "Running PostgreSQL migrations..."
    
    # Create migrations table if it doesn't exist
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB << EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64),
    execution_time_ms INTEGER
);
EOF

    # Get list of applied migrations
    applied_migrations=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT version FROM schema_migrations ORDER BY version;" | tr -d ' ')
    
    # Run pending migrations
    for migration_file in $(ls migrations/*.sql | sort); do
        migration_version=$(basename $migration_file .sql)
        
        if ! echo "$applied_migrations" | grep -q "^$migration_version$"; then
            log_info "Applying migration: $migration_version"
            
            # Calculate checksum
            checksum=$(sha256sum $migration_file | cut -d' ' -f1)
            
            # Measure execution time
            start_time=$(date +%s%3N)
            PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -f $migration_file
            end_time=$(date +%s%3N)
            execution_time=$((end_time - start_time))
            
            # Record applied migration
            PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "INSERT INTO schema_migrations (version, checksum, execution_time_ms) VALUES ('$migration_version', '$checksum', $execution_time);"
            log_info "Migration $migration_version applied successfully (${execution_time}ms)"
        else
            log_debug "Migration $migration_version already applied, skipping"
        fi
    done
}

# Rollback last migration
rollback_postgres_migration() {
    log_info "Rolling back last PostgreSQL migration..."
    
    # Get last applied migration
    last_migration=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;" | tr -d ' ')
    
    if [ -z "$last_migration" ]; then
        log_warn "No migrations to rollback"
        return 0
    fi
    
    rollback_file="migrations/${last_migration}_rollback.sql"
    if [ -f "$rollback_file" ]; then
        log_info "Rolling back migration: $last_migration"
        PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -f $rollback_file
        PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DELETE FROM schema_migrations WHERE version = '$last_migration';"
        log_info "Migration $last_migration rolled back successfully"
    else
        log_error "Rollback file not found: $rollback_file"
        return 1
    fi
}

# Check if MongoDB is available
check_mongo() {
    log_info "Checking MongoDB connection..."
    if mongosh --host $MONGO_HOST:$MONGO_PORT --eval "db.runCommand('ping')" > /dev/null 2>&1; then
        log_info "MongoDB connection successful"
        return 0
    else
        log_error "Failed to connect to MongoDB"
        return 1
    fi
}

# Setup MongoDB
setup_mongo() {
    log_info "Setting up MongoDB..."
    
    # Run MongoDB initialization scripts
    for init_file in $(ls mongo-init/*.js 2>/dev/null | sort); do
        log_info "Running MongoDB script: $(basename $init_file)"
        mongosh --host $MONGO_HOST:$MONGO_PORT $MONGO_DB --eval "load('$init_file')"
    done
}

# Check if Redis is available
check_redis() {
    log_info "Checking Redis connection..."
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping > /dev/null 2>&1; then
        log_info "Redis connection successful"
        return 0
    else
        log_error "Failed to connect to Redis"
        return 1
    fi
}

# Create tenant schema
create_tenant_schema() {
    local tenant_slug=$1
    if [ -z "$tenant_slug" ]; then
        log_error "Tenant slug is required"
        return 1
    fi
    
    log_info "Creating schema for tenant: $tenant_slug"
    
    # Create PostgreSQL schema
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT create_tenant_schema('$tenant_slug');"
    
    # Create MongoDB database
    mongosh --host $MONGO_HOST:$MONGO_PORT "tenant_${tenant_slug}" --eval "
        load('mongo-init/tenant-collections.js');
        print('Tenant MongoDB database created: tenant_${tenant_slug}');
    "
    
    log_info "Tenant schema created successfully for: $tenant_slug"
}

# Drop tenant schema
drop_tenant_schema() {
    local tenant_slug=$1
    if [ -z "$tenant_slug" ]; then
        log_error "Tenant slug is required"
        return 1
    fi
    
    log_warn "Dropping schema for tenant: $tenant_slug"
    read -p "Are you sure you want to drop all data for tenant '$tenant_slug'? (yes/no): " confirm
    
    if [ "$confirm" = "yes" ]; then
        # Drop PostgreSQL schema
        PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT drop_tenant_schema('$tenant_slug');"
        
        # Drop MongoDB database
        mongosh --host $MONGO_HOST:$MONGO_PORT "tenant_${tenant_slug}" --eval "db.dropDatabase();"
        
        # Clean Redis keys
        redis-cli -h $REDIS_HOST -p $REDIS_PORT EVAL "
            local keys = redis.call('keys', ARGV[1])
            if #keys > 0 then
                return redis.call('del', unpack(keys))
            end
            return 0
        " 0 "${tenant_slug}:*"
        
        log_info "Tenant schema dropped successfully for: $tenant_slug"
    else
        log_info "Operation cancelled"
    fi
}

# List migrations status
migration_status() {
    log_info "Migration Status:"
    echo "===================="
    
    # Get applied migrations
    applied_migrations=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "
        SELECT 
            version, 
            applied_at, 
            execution_time_ms || 'ms' as execution_time
        FROM schema_migrations 
        ORDER BY version;
    " 2>/dev/null || echo "No migrations table found")
    
    echo "$applied_migrations"
    echo ""
    
    # Check for pending migrations
    all_migrations=$(ls migrations/*.sql | wc -l)
    applied_count=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT COUNT(*) FROM schema_migrations;" 2>/dev/null || echo "0")
    pending_count=$((all_migrations - applied_count))
    
    log_info "Total migrations: $all_migrations"
    log_info "Applied migrations: $applied_count"
    log_info "Pending migrations: $pending_count"
}

# Validate database schemas
validate_schemas() {
    log_info "Validating database schemas..."
    
    # Check PostgreSQL tables
    tables=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -t -c "
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_type = 'BASE TABLE'
        ORDER BY table_name;
    ")
    
    log_info "PostgreSQL tables found:"
    echo "$tables"
    
    # Check MongoDB collections
    if check_mongo; then
        log_info "MongoDB databases:"
        mongosh --host $MONGO_HOST:$MONGO_PORT --eval "db.adminCommand('listDatabases').databases.forEach(function(db) { print(db.name); });"
    fi
}

# Main function
main() {
    case "${1:-all}" in
        "postgres")
            if check_postgres; then
                run_postgres_migrations
            else
                exit 1
            fi
            ;;
        "mongo")
            if check_mongo; then
                setup_mongo
            else
                exit 1
            fi
            ;;
        "redis")
            check_redis
            ;;
        "rollback")
            if check_postgres; then
                rollback_postgres_migration
            else
                exit 1
            fi
            ;;
        "status")
            migration_status
            ;;
        "validate")
            validate_schemas
            ;;
        "create-tenant")
            if [ -z "$2" ]; then
                log_error "Usage: $0 create-tenant <tenant_slug>"
                exit 1
            fi
            create_tenant_schema "$2"
            ;;
        "drop-tenant")
            if [ -z "$2" ]; then
                log_error "Usage: $0 drop-tenant <tenant_slug>"
                exit 1
            fi
            drop_tenant_schema "$2"
            ;;
        "all")
            if check_postgres; then
                run_postgres_migrations
            else
                log_error "PostgreSQL migration failed"
                exit 1
            fi
            
            if check_mongo; then
                setup_mongo
            else
                log_warn "MongoDB setup failed, continuing..."
            fi
            
            if ! check_redis; then
                log_warn "Redis connection failed, continuing..."
            fi
            ;;
        *)
            echo "Usage: $0 {postgres|mongo|redis|rollback|status|validate|create-tenant|drop-tenant|all}"
            echo ""
            echo "Commands:"
            echo "  postgres              - Run PostgreSQL migrations only"
            echo "  mongo                 - Setup MongoDB only"
            echo "  redis                 - Check Redis connection only"
            echo "  rollback              - Rollback last PostgreSQL migration"
            echo "  status                - Show migration status"
            echo "  validate              - Validate database schemas"
            echo "  create-tenant <slug>  - Create schema for new tenant"
            echo "  drop-tenant <slug>    - Drop schema for tenant (DANGEROUS)"
            echo "  all                   - Run all database setups (default)"
            exit 1
            ;;
    esac
    
    log_info "Database operation completed successfully!"
}

main "$@"

set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-zplus_saas_base}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database connection string
DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

# Directories
MIGRATIONS_DIR="$(dirname "$0")/migrations"
SEEDERS_DIR="$(dirname "$0")/seeders"

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

check_connection() {
    log_info "Checking database connection..."
    if psql "$DB_URL" -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "Database connection successful"
        return 0
    else
        log_error "Failed to connect to database"
        log_error "Connection string: $DB_URL"
        return 1
    fi
}

create_migration_table() {
    log_info "Creating migration tracking table..."
    psql "$DB_URL" -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    " > /dev/null
    log_success "Migration table created"
}

run_migrations() {
    log_info "Running database migrations..."
    
    # Create migration table if it doesn't exist
    create_migration_table
    
    # Get list of applied migrations
    applied_migrations=$(psql "$DB_URL" -t -c "SELECT version FROM schema_migrations ORDER BY version;")
    
    # Run each migration file
    for migration_file in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$migration_file" ]; then
            filename=$(basename "$migration_file")
            version="${filename%.*}"
            
            # Check if migration already applied
            if echo "$applied_migrations" | grep -q "$version"; then
                log_warning "Migration $version already applied, skipping..."
                continue
            fi
            
            log_info "Applying migration: $version"
            
            # Run migration
            if psql "$DB_URL" -f "$migration_file" > /dev/null; then
                # Record migration as applied
                psql "$DB_URL" -c "INSERT INTO schema_migrations (version) VALUES ('$version');" > /dev/null
                log_success "Migration $version applied successfully"
            else
                log_error "Failed to apply migration: $version"
                exit 1
            fi
        fi
    done
    
    log_success "All migrations completed"
}

run_seeders() {
    log_info "Running database seeders..."
    
    # Check if migrations have been run
    if ! psql "$DB_URL" -c "SELECT 1 FROM schema_migrations LIMIT 1;" > /dev/null 2>&1; then
        log_error "No migrations found. Please run migrations first."
        exit 1
    fi
    
    # Run each seeder file
    for seeder_file in "$SEEDERS_DIR"/*.sql; do
        if [ -f "$seeder_file" ]; then
            filename=$(basename "$seeder_file")
            seeder_name="${filename%.*}"
            
            log_info "Running seeder: $seeder_name"
            
            if psql "$DB_URL" -f "$seeder_file" > /dev/null; then
                log_success "Seeder $seeder_name completed successfully"
            else
                log_error "Failed to run seeder: $seeder_name"
                exit 1
            fi
        fi
    done
    
    log_success "All seeders completed"
}

reset_database() {
    log_warning "Resetting database (this will DELETE ALL DATA)..."
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Database reset cancelled"
        exit 0
    fi
    
    log_info "Dropping all tables..."
    
    # Drop all tables
    psql "$DB_URL" -c "
        DROP TABLE IF EXISTS rate_limit_logs CASCADE;
        DROP TABLE IF EXISTS api_keys CASCADE;
        DROP TABLE IF EXISTS webhooks CASCADE;
        DROP TABLE IF EXISTS audit_logs CASCADE;
        DROP TABLE IF EXISTS user_sessions CASCADE;
        DROP TABLE IF EXISTS users CASCADE;
        DROP TABLE IF EXISTS tenant_domains CASCADE;
        DROP TABLE IF EXISTS tenants CASCADE;
        DROP TABLE IF EXISTS schema_migrations CASCADE;
        DROP FUNCTION IF EXISTS update_updated_at_column CASCADE;
    " > /dev/null
    
    log_success "Database reset completed"
    
    # Run migrations and seeders
    run_migrations
    run_seeders
}

show_status() {
    log_info "Database Status"
    echo "===================="
    
    # Check connection
    if ! check_connection; then
        return 1
    fi
    
    # Show migration status
    echo
    log_info "Migration Status:"
    if psql "$DB_URL" -c "SELECT version, applied_at FROM schema_migrations ORDER BY applied_at;" 2>/dev/null; then
        echo
    else
        log_warning "No migrations applied yet"
        echo
    fi
    
    # Show table counts
    log_info "Table Statistics:"
    psql "$DB_URL" -c "
        SELECT 
            schemaname,
            tablename,
            n_tup_ins as inserts,
            n_tup_upd as updates,
            n_tup_del as deletes,
            n_live_tup as live_rows
        FROM pg_stat_user_tables 
        ORDER BY tablename;
    " 2>/dev/null || log_warning "Could not retrieve table statistics"
}

# Main script logic
case "${1:-help}" in
    migrate)
        check_connection || exit 1
        run_migrations
        ;;
    seed)
        check_connection || exit 1
        run_seeders
        ;;
    reset)
        check_connection || exit 1
        reset_database
        ;;
    status)
        show_status
        ;;
    help|*)
        echo "Database Migration and Seeder Script"
        echo
        echo "Usage: $0 [command]"
        echo
        echo "Commands:"
        echo "  migrate    Run pending database migrations"
        echo "  seed       Run database seeders"
        echo "  reset      Reset database (DROP ALL DATA and re-create)"
        echo "  status     Show database and migration status"
        echo "  help       Show this help message"
        echo
        echo "Environment Variables:"
        echo "  DB_HOST     Database host (default: localhost)"
        echo "  DB_PORT     Database port (default: 5432)"
        echo "  DB_NAME     Database name (default: zplus_saas_base)"
        echo "  DB_USER     Database user (default: postgres)"
        echo "  DB_PASSWORD Database password (default: postgres)"
        ;;
esac
