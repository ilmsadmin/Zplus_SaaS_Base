#!/bin/bash

# Advanced Database Management Script for Zplus SaaS Base
# Handles backup, restore, monitoring, and maintenance tasks

set -e

# Configuration from environment variables
POSTGRES_HOST=${POSTGRES_HOST:-localhost}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_DB=${POSTGRES_DB:-zplus_saas}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-password}

MONGO_HOST=${MONGO_HOST:-localhost}
MONGO_PORT=${MONGO_PORT:-27017}

REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}

BACKUP_DIR=${BACKUP_DIR:-./backups}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="docker-compose.database.yml"
PROJECT_NAME="zplus"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_header() { echo -e "${PURPLE}[ZPLUS DB]${NC} $1"; }

show_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════╗"
    echo "║          Zplus SaaS Database Manager         ║"
    echo "╚══════════════════════════════════════════════╝"
    echo -e "${NC}"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
}

wait_for_services() {
    log_info "Waiting for database services to be healthy..."
    
    local max_attempts=60
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f "$COMPOSE_FILE" ps | grep -q "healthy"; then
            local healthy_count=$(docker-compose -f "$COMPOSE_FILE" ps | grep -c "healthy" || echo "0")
            local total_db_services=3  # postgres, mongodb, redis
            
            if [ "$healthy_count" -ge "$total_db_services" ]; then
                log_success "All database services are healthy!"
                return 0
            fi
        fi
        
        echo -n "."
        sleep 2
        ((attempt++))
    done
    
    log_error "Timeout waiting for services to be healthy"
    return 1
}

start_databases() {
    log_header "Starting Database Services"
    
    if docker-compose -f "$COMPOSE_FILE" up -d; then
        log_success "Database containers started"
        wait_for_services
    else
        log_error "Failed to start database containers"
        exit 1
    fi
}

stop_databases() {
    log_header "Stopping Database Services"
    
    if docker-compose -f "$COMPOSE_FILE" down; then
        log_success "Database containers stopped"
    else
        log_error "Failed to stop database containers"
        exit 1
    fi
}

restart_databases() {
    log_header "Restarting Database Services"
    stop_databases
    start_databases
}

setup_databases() {
    log_header "Setting Up Databases"
    
    # Start services first
    start_databases
    
    # Run migrations
    log_info "Running database migrations..."
    cd backend/database
    ./db-migrate.sh migrate
    
    # Run seeders
    log_info "Running database seeders..."
    ./db-migrate.sh seed
    
    cd ../..
    log_success "Database setup completed!"
}

show_status() {
    log_header "Database Services Status"
    
    echo
    log_info "Container Status:"
    docker-compose -f "$COMPOSE_FILE" ps
    
    echo
    log_info "Database Connection Status:"
    
    # Test PostgreSQL
    if docker exec zplus_postgres pg_isready -U postgres -d zplus_saas_base &>/dev/null; then
        log_success "PostgreSQL: Connected"
    else
        log_error "PostgreSQL: Connection failed"
    fi
    
    # Test MongoDB
    if docker exec zplus_mongodb mongosh --eval "db.adminCommand('ping')" &>/dev/null; then
        log_success "MongoDB: Connected"
    else
        log_error "MongoDB: Connection failed"
    fi
    
    # Test Redis
    if docker exec zplus_redis redis-cli -a redis123 ping &>/dev/null; then
        log_success "Redis: Connected"
    else
        log_error "Redis: Connection failed"
    fi
    
    echo
    log_info "Migration Status:"
    cd backend/database
    ./db-migrate.sh status || true
    cd ../..
}

show_logs() {
    local service="${1:-}"
    
    if [ -n "$service" ]; then
        log_info "Showing logs for service: $service"
        docker-compose -f "$COMPOSE_FILE" logs -f "$service"
    else
        log_info "Showing logs for all database services"
        docker-compose -f "$COMPOSE_FILE" logs -f
    fi
}

backup_databases() {
    log_header "Creating Database Backup"
    
    # Create backup directory
    mkdir -p ./backups
    
    # PostgreSQL backup
    log_info "Backing up PostgreSQL..."
    docker exec zplus_postgres /scripts/backup.sh
    
    # MongoDB backup
    log_info "Backing up MongoDB..."
    docker exec zplus_mongodb mongodump --host localhost --out /tmp/mongo_backup_$(date +%Y%m%d_%H%M%S)
    
    log_success "Database backup completed!"
}

open_management_tools() {
    log_header "Database Management Tools"
    
    echo
    echo "Available management interfaces:"
    echo "┌─────────────────────────────────────────────────────────┐"
    echo "│ PostgreSQL (pgAdmin)     │ http://localhost:8080        │"
    echo "│ MongoDB (Mongo Express)  │ http://localhost:8081        │"
    echo "│ Redis (Redis Commander)  │ http://localhost:8082        │"
    echo "│ Universal (Adminer)      │ http://localhost:8083        │"
    echo "└─────────────────────────────────────────────────────────┘"
    echo
    
    # Try to open in browser (macOS)
    if command -v open &> /dev/null; then
        read -p "Open pgAdmin in browser? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            open http://localhost:8080
        fi
    fi
}

reset_databases() {
    log_warning "This will DELETE ALL DATABASE DATA!"
    read -p "Are you sure you want to continue? (type 'yes' to confirm): " -r
    echo
    
    if [[ ! $REPLY == "yes" ]]; then
        log_info "Reset cancelled"
        return 0
    fi
    
    log_header "Resetting Databases"
    
    # Stop services
    stop_databases
    
    # Remove volumes
    log_info "Removing database volumes..."
    docker volume rm ${PROJECT_NAME}_postgres_data 2>/dev/null || true
    docker volume rm ${PROJECT_NAME}_mongodb_data 2>/dev/null || true
    docker volume rm ${PROJECT_NAME}_redis_data 2>/dev/null || true
    docker volume rm ${PROJECT_NAME}_pgadmin_data 2>/dev/null || true
    
    # Setup fresh databases
    setup_databases
    
    log_success "Database reset completed!"
}

monitor_performance() {
    log_header "Database Performance Monitor"
    
    echo "Real-time database performance monitoring"
    echo "Press Ctrl+C to stop"
    echo
    
    while true; do
        clear
        echo -e "${CYAN}=== Database Performance Monitor ===${NC}"
        echo "Updated: $(date)"
        echo
        
        # PostgreSQL stats
        echo -e "${YELLOW}PostgreSQL:${NC}"
        docker exec zplus_postgres psql -U postgres -d zplus_saas_base -c "
            SELECT 
                'Connections' as metric, 
                numbackends as value 
            FROM pg_stat_database 
            WHERE datname = 'zplus_saas_base'
            UNION ALL
            SELECT 
                'Transactions/sec' as metric, 
                ROUND((xact_commit + xact_rollback)::numeric, 2) as value
            FROM pg_stat_database 
            WHERE datname = 'zplus_saas_base';
        " 2>/dev/null || echo "  Connection failed"
        
        echo
        
        # MongoDB stats
        echo -e "${YELLOW}MongoDB:${NC}"
        docker exec zplus_mongodb mongosh --quiet --eval "
            const stats = db.runCommand({serverStatus: 1});
            print('  Connections: ' + stats.connections.current);
            print('  Operations/sec: ' + (stats.opcounters.query + stats.opcounters.insert + stats.opcounters.update + stats.opcounters.delete));
        " 2>/dev/null || echo "  Connection failed"
        
        echo
        
        # Redis stats
        echo -e "${YELLOW}Redis:${NC}"
        docker exec zplus_redis redis-cli -a redis123 --csv INFO stats 2>/dev/null | grep -E "(connected_clients|total_commands_processed)" | sed 's/,/: /g' | sed 's/^/  /' || echo "  Connection failed"
        
        sleep 5
    done
}

show_help() {
    show_banner
    echo "Database management script for Zplus SaaS Base"
    echo
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  start      Start all database services"
    echo "  stop       Stop all database services"
    echo "  restart    Restart all database services"
    echo "  setup      Initial database setup (start + migrate + seed)"
    echo "  status     Show database services status"
    echo "  logs       Show database logs [service_name]"
    echo "  backup     Create database backup"
    echo "  reset      Reset all databases (DELETE ALL DATA)"
    echo "  tools      Show database management tools URLs"
    echo "  monitor    Real-time performance monitoring"
    echo "  help       Show this help message"
    echo
    echo "Examples:"
    echo "  $0 setup              # Initial setup"
    echo "  $0 logs postgres      # Show PostgreSQL logs"
    echo "  $0 status             # Check service status"
    echo
}

# Main script logic
check_docker

case "${1:-help}" in
    start)
        start_databases
        ;;
    stop)
        stop_databases
        ;;
    restart)
        restart_databases
        ;;
    setup)
        setup_databases
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$2"
        ;;
    backup)
        backup_databases
        ;;
    reset)
        reset_databases
        ;;
    tools)
        open_management_tools
        ;;
    monitor)
        monitor_performance
        ;;
    help|*)
        show_help
        ;;
esac
