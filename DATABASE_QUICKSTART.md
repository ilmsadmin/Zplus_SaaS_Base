# Quick Start - Database Setup

## 🚀 Get Started in 3 Steps

### Step 1: Start Database Services
```bash
make db-dev
```

### Step 2: Access Management Tools
- **pgAdmin (PostgreSQL)**: http://localhost:8080
  - Email: `admin@zplus.io`
  - Password: `pgadmin123`

- **Mongo Express (MongoDB)**: http://localhost:8081
  - Username: `admin`
  - Password: `mongoexpress123`

- **Redis Commander**: http://localhost:8082
  - Username: `admin`
  - Password: `rediscommander123`

- **Adminer (Universal)**: http://localhost:8083
  - Server: `postgres`
  - Username: `postgres`
  - Password: `postgres123`

### Step 3: Explore Sample Data
The database comes pre-loaded with sample data including:

#### Sample Tenants
- **demo** - Demo Corporation (Professional plan)
- **acme** - ACME Inc (Enterprise plan)
- **startup** - Startup Hub (Starter plan)
- **enterprise** - Enterprise Solutions (Enterprise plan)

#### Sample Users
- **System Admin**: `admin@zplus.io`
- **Demo Admin**: `admin@demo.com`
- **ACME Admin**: `admin@acme.com`
- Multiple regular users across tenants

#### Sample Data Features
- ✅ Multi-tenant isolation
- ✅ Custom domain support
- ✅ API keys with rate limiting
- ✅ Webhooks for integrations
- ✅ Audit logs
- ✅ Role-based access control

## 🔧 Available Commands

```bash
# Database management
make db-start      # Start databases
make db-stop       # Stop databases
make db-status     # Check status
make db-logs       # View logs
make db-backup     # Create backup
make db-reset      # Reset (WARNING: deletes data)

# Development
make db-dev        # Quick development setup
make db-tools      # Show management URLs
make db-monitor    # Performance monitoring
```

## 📊 Sample Queries

### PostgreSQL (via pgAdmin)
```sql
-- View all tenants with their domains
SELECT t.name, t.plan, td.domain, td.is_custom 
FROM tenants t 
LEFT JOIN tenant_domains td ON t.id = td.tenant_id;

-- Check API key usage
SELECT ak.name, t.name as tenant, ak.usage_count, ak.last_used_at
FROM api_keys ak 
JOIN tenants t ON ak.tenant_id = t.id
ORDER BY ak.usage_count DESC;
```

### MongoDB (via Mongo Express)
Navigate to `zplus_saas_base` database and explore:
- `system_settings` - Global configuration
- `files` - File metadata
- `application_logs` - Application logs

For tenant-specific data, check `tenant_demo`, `tenant_acme`, etc.

### Redis (via Redis Commander)
Explore key patterns:
- `session:*` - User sessions
- `cache:tenant:*` - Tenant data cache
- `domain_mapping:*` - Domain to tenant mapping

## 🛠️ Troubleshooting

### Services won't start?
```bash
# Check Docker status
docker ps

# View specific service logs
make db-logs service=postgres

# Restart services
make db-restart
```

### Can't connect to management tools?
1. Ensure all services are running: `make db-status`
2. Check ports 8080-8083 are not blocked
3. Try accessing via http://127.0.0.1:8080 instead of localhost

### Need to reset everything?
```bash
# WARNING: This deletes all data
make db-reset
```

---

**Ready to start coding?** Check out the [Database Management Guide](../docs/database/Database_Management_Guide.md) for advanced usage!
