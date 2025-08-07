# Database Management Guide - Visual Interface Access

## Overview

Zplus SaaS Base provides multiple visual database management tools for easy database administration and monitoring. This guide covers how to access and use each tool.

## 🚀 Quick Start

### Start Database Services

```bash
# Start all database services and management tools
docker-compose -f docker-compose.database.yml up -d

# Check service status
docker-compose -f docker-compose.database.yml ps

# View logs
docker-compose -f docker-compose.database.yml logs -f
```

### Run Migrations and Seeders

```bash
# Navigate to database directory
cd backend/database

# Set environment variables (optional, defaults provided)
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=zplus_saas_base
export DB_USER=postgres
export DB_PASSWORD=postgres123

# Run migrations
./db-migrate.sh migrate

# Run seeders
./db-migrate.sh seed

# Check status
./db-migrate.sh status
```

## 🛠️ Database Management Tools

### 1. pgAdmin (PostgreSQL) 
**Access:** http://localhost:8080

#### Login Credentials
- **Email:** admin@zplus.io
- **Password:** pgadmin123

#### Features
- Complete PostgreSQL database administration
- Query editor with syntax highlighting
- Database schema visualization
- Backup and restore functionality
- Performance monitoring
- User and permission management

#### Quick Actions
1. **Connect to Database:**
   - Server already configured as "Zplus PostgreSQL"
   - Password: `postgres123`

2. **Browse Tables:**
   - Navigate: Servers → Zplus PostgreSQL → Databases → zplus_saas_base → Schemas → public → Tables

3. **Run Queries:**
   - Right-click database → Query Tool
   - Execute custom SQL queries

4. **View Data:**
   - Right-click table → View/Edit Data → All Rows

### 2. Mongo Express (MongoDB)
**Access:** http://localhost:8081

#### Login Credentials
- **Username:** admin
- **Password:** mongoexpress123

#### Features
- MongoDB collection browsing
- Document viewing and editing
- Index management
- GridFS file browser
- Database statistics

#### Quick Actions
1. **Browse Collections:**
   - Select database from dropdown
   - Click on collection name to view documents

2. **View Documents:**
   - Click on document ID to view/edit
   - Use JSON editor for modifications

3. **Create New Document:**
   - Click "New Document" in collection view
   - Enter JSON data

### 3. Redis Commander (Redis)
**Access:** http://localhost:8082

#### Login Credentials
- **Username:** admin
- **Password:** rediscommander123

#### Features
- Redis key-value browser
- Real-time monitoring
- Memory usage statistics
- CLI interface
- Key expiration management

#### Quick Actions
1. **Browse Keys:**
   - Use tree view to navigate key namespaces
   - Click on key to view value

2. **Monitor Performance:**
   - View real-time statistics
   - Monitor memory usage
   - Check connected clients

3. **Execute Commands:**
   - Use built-in CLI interface
   - Run Redis commands directly

### 4. Adminer (Universal Database Tool)
**Access:** http://localhost:8083

#### Login Credentials
- **System:** PostgreSQL
- **Server:** postgres
- **Username:** postgres
- **Password:** postgres123
- **Database:** zplus_saas_base

#### Features
- Lightweight database administration
- Support for multiple database types
- SQL query execution
- Database structure overview
- Data import/export

#### Quick Actions
1. **Login:**
   - Select PostgreSQL system
   - Enter credentials above

2. **Browse Tables:**
   - Click on table names in left sidebar
   - View table structure and data

3. **Execute SQL:**
   - Use SQL command interface
   - Run custom queries

## 📊 Database Overview

### PostgreSQL Tables

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `tenants` | Store tenant information | Multi-tenant isolation, plan management |
| `tenant_domains` | Custom domain management | SSL support, domain verification |
| `users` | User accounts | Role-based access, Keycloak integration |
| `user_sessions` | Session management | Security tracking, device management |
| `api_keys` | API access keys | Rate limiting, scope-based permissions |
| `webhooks` | External integrations | Event-driven notifications |
| `audit_logs` | Activity tracking | Comprehensive audit trail |
| `rate_limit_logs` | API rate limiting | Usage monitoring and throttling |

### MongoDB Collections

| Collection | Purpose | Database |
|------------|---------|----------|
| `system_settings` | Global configuration | zplus_saas_base |
| `files` | File metadata | zplus_saas_base |
| `application_logs` | Application logs | zplus_saas_base |
| `documents` | Tenant documents | tenant_{tenant_id} |

### Redis Key Patterns

| Pattern | Purpose | Example |
|---------|---------|---------|
| `session:{token}` | User sessions | `session:abc123...` |
| `cache:tenant:{id}` | Tenant data cache | `cache:tenant:demo` |
| `rate_limit:{key}:{window}` | Rate limiting | `rate_limit:api_xyz:2025010115` |
| `domain_mapping:{domain}` | Domain to tenant mapping | `domain_mapping:app.acme.com` |

## 🔍 Sample Queries

### PostgreSQL Queries

```sql
-- View all tenants with their domains
SELECT 
    t.id,
    t.name,
    t.plan,
    t.status,
    td.domain,
    td.is_custom,
    td.verified
FROM tenants t
LEFT JOIN tenant_domains td ON t.id = td.tenant_id
ORDER BY t.created_at DESC;

-- Check user distribution by tenant
SELECT 
    t.name as tenant_name,
    COUNT(u.id) as user_count,
    t.max_users,
    ROUND((COUNT(u.id)::float / t.max_users) * 100, 2) as usage_percentage
FROM tenants t
LEFT JOIN users u ON t.id = u.tenant_id AND u.deleted_at IS NULL
GROUP BY t.id, t.name, t.max_users
ORDER BY usage_percentage DESC;

-- API key usage statistics
SELECT 
    ak.name,
    t.name as tenant,
    ak.rate_limit_per_hour,
    ak.usage_count,
    ak.last_used_at,
    CASE WHEN ak.expires_at < NOW() THEN 'Expired'
         WHEN ak.revoked THEN 'Revoked'
         ELSE 'Active' END as status
FROM api_keys ak
JOIN tenants t ON ak.tenant_id = t.id
ORDER BY ak.usage_count DESC;

-- Recent audit activities
SELECT 
    al.action,
    al.resource_type,
    u.email as user_email,
    t.name as tenant_name,
    al.created_at
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
LEFT JOIN tenants t ON al.tenant_id = t.id
ORDER BY al.created_at DESC
LIMIT 50;
```

### MongoDB Queries

```javascript
// Find all system settings
db.system_settings.find().pretty()

// Get file statistics by tenant
db.files.aggregate([
  {
    $group: {
      _id: "$tenant_id",
      total_files: { $sum: 1 },
      total_size: { $sum: "$file_size" },
      avg_size: { $avg: "$file_size" }
    }
  },
  {
    $sort: { total_size: -1 }
  }
])

// Recent application logs
db.application_logs.find({
  timestamp: {
    $gte: new Date(Date.now() - 24*60*60*1000) // Last 24 hours
  }
}).sort({ timestamp: -1 }).limit(100)

// Search documents across tenant
// (switch to tenant database first: use tenant_demo)
db.documents.find({
  $text: { $search: "sample" }
}).limit(10)
```

### Redis Commands

```bash
# View all keys
KEYS *

# Get domain mapping
GET domain_mapping:app.acme.com

# Check session
GET session:your_session_token

# View rate limiting
KEYS rate_limit:*

# Monitor real-time commands
MONITOR

# Get memory info
INFO memory

# View connected clients
CLIENT LIST
```

## 🔧 Troubleshooting

### Common Issues

#### 1. Container Won't Start
```bash
# Check logs
docker-compose -f docker-compose.database.yml logs [service_name]

# Restart specific service
docker-compose -f docker-compose.database.yml restart [service_name]

# Remove and recreate
docker-compose -f docker-compose.database.yml down
docker-compose -f docker-compose.database.yml up -d
```

#### 2. Database Connection Failed
```bash
# Check if database is running
docker-compose -f docker-compose.database.yml ps

# Test connection manually
docker exec -it zplus_postgres psql -U postgres -d zplus_saas_base -c "SELECT 1;"
```

#### 3. Migration Errors
```bash
# Check migration status
./db-migrate.sh status

# Reset database (WARNING: Deletes all data)
./db-migrate.sh reset
```

#### 4. Access Issues
- Ensure all services are running: `docker-compose -f docker-compose.database.yml ps`
- Check firewall settings for ports 8080-8083
- Verify credentials in the guide above

### Performance Monitoring

#### PostgreSQL Performance
```sql
-- Check slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Check table sizes
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY size_bytes DESC;
```

#### MongoDB Performance
```javascript
// Check slow operations
db.runCommand({profile: 2, slowms: 100})
db.system.profile.find().limit(5).sort({ts: -1}).pretty()

// Collection statistics
db.stats()
db.files.stats()
```

#### Redis Performance
```bash
# Performance info
INFO stats
INFO replication
INFO memory

# Slow log
SLOWLOG GET 10
```

## 🛡️ Security Notes

### Default Credentials (Development Only)
**⚠️ Important:** Change all default passwords before production deployment!

- PostgreSQL: `postgres` / `postgres123`
- MongoDB: `admin` / `mongodb123`
- Redis: Password `redis123`
- pgAdmin: `admin@zplus.io` / `pgadmin123`
- Mongo Express: `admin` / `mongoexpress123`
- Redis Commander: `admin` / `rediscommander123`

### Production Security Checklist
- [ ] Change all default passwords
- [ ] Enable SSL/TLS for all connections
- [ ] Configure firewall rules
- [ ] Enable audit logging
- [ ] Set up database backup encryption
- [ ] Configure access restrictions by IP
- [ ] Enable two-factor authentication where available
- [ ] Regular security updates

## 📋 Backup and Restore

### Automated Backups
The `db-backup` service automatically creates daily PostgreSQL backups in the `./backups` directory.

### Manual Backup Commands

#### PostgreSQL
```bash
# Full backup
docker exec zplus_postgres pg_dump -U postgres zplus_saas_base > backup_$(date +%Y%m%d).sql

# Restore
docker exec -i zplus_postgres psql -U postgres zplus_saas_base < backup_20250807.sql
```

#### MongoDB
```bash
# Backup
docker exec zplus_mongodb mongodump --host localhost --out /tmp/backup

# Restore
docker exec zplus_mongodb mongorestore --host localhost /tmp/backup
```

#### Redis
```bash
# Backup (RDB snapshot)
docker exec zplus_redis redis-cli -a redis123 BGSAVE

# Copy backup file
docker cp zplus_redis:/data/dump.rdb ./redis_backup.rdb
```

---

**Need Help?** Check the troubleshooting section or review container logs for specific error messages.
