#!/bin/bash

# Database Migration and Seeder Script
# Usage: ./db-migrate.sh [migrate|seed|reset|status]

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
