#!/bin/bash

# Backup script for PostgreSQL database
# This script creates compressed backups with rotation

set -e

# Configuration
BACKUP_DIR="/backups"
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-zplus_saas_base}"
DB_USER="${DB_USER:-postgres}"
PGPASSWORD="${PGPASSWORD:-postgres123}"

# Backup retention (days)
RETENTION_DAYS=30

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/zplus_backup_$TIMESTAMP.sql"
COMPRESSED_FILE="$BACKUP_FILE.gz"

log_info "Starting database backup..."
log_info "Database: $DB_NAME"
log_info "Host: $DB_HOST:$DB_PORT"
log_info "Backup file: $COMPRESSED_FILE"

# Create backup
if pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
   --verbose \
   --no-password \
   --format=custom \
   --no-privileges \
   --no-tablespaces \
   --compress=9 > "$BACKUP_FILE.custom" 2>/dev/null; then
    
    log_success "Database backup created: $BACKUP_FILE.custom"
    
    # Also create SQL dump for easy viewing
    pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
       --verbose \
       --no-password \
       --format=plain \
       --no-privileges \
       --no-tablespaces | gzip > "$COMPRESSED_FILE" 2>/dev/null
    
    log_success "SQL dump created: $COMPRESSED_FILE"
    
    # Get file sizes
    CUSTOM_SIZE=$(du -h "$BACKUP_FILE.custom" | cut -f1)
    SQL_SIZE=$(du -h "$COMPRESSED_FILE" | cut -f1)
    
    log_info "Backup sizes - Custom: $CUSTOM_SIZE, SQL: $SQL_SIZE"
    
else
    log_error "Failed to create database backup"
    exit 1
fi

# Remove old backups
log_info "Cleaning up old backups (older than $RETENTION_DAYS days)..."

find "$BACKUP_DIR" -name "zplus_backup_*.sql*" -type f -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "zplus_backup_*.custom" -type f -mtime +$RETENTION_DAYS -delete 2>/dev/null || true

REMAINING_BACKUPS=$(find "$BACKUP_DIR" -name "zplus_backup_*" -type f | wc -l)
log_info "Backup cleanup completed. Remaining backups: $REMAINING_BACKUPS"

# Verify backup integrity
log_info "Verifying backup integrity..."
if pg_restore --list "$BACKUP_FILE.custom" > /dev/null 2>&1; then
    log_success "Backup integrity verified"
else
    log_warning "Backup integrity check failed"
fi

# Create backup manifest
MANIFEST_FILE="$BACKUP_DIR/backup_manifest.json"
cat > "$MANIFEST_FILE" << EOF
{
  "latest_backup": {
    "timestamp": "$TIMESTAMP",
    "database": "$DB_NAME",
    "files": {
      "custom_format": "$(basename "$BACKUP_FILE.custom")",
      "sql_format": "$(basename "$COMPRESSED_FILE")"
    },
    "sizes": {
      "custom_format": "$CUSTOM_SIZE",
      "sql_format": "$SQL_SIZE"
    },
    "created_at": "$(date -Iseconds)"
  },
  "retention_days": $RETENTION_DAYS,
  "total_backups": $REMAINING_BACKUPS
}
EOF

log_success "Backup manifest updated: $MANIFEST_FILE"
log_success "Database backup completed successfully!"

# Optional: Send notification (webhook, email, etc.)
if [ -n "$BACKUP_WEBHOOK_URL" ]; then
    curl -X POST "$BACKUP_WEBHOOK_URL" \
         -H "Content-Type: application/json" \
         -d "{\"status\":\"success\",\"database\":\"$DB_NAME\",\"timestamp\":\"$TIMESTAMP\",\"size\":\"$CUSTOM_SIZE\"}" \
         > /dev/null 2>&1 || log_warning "Failed to send backup notification"
fi
