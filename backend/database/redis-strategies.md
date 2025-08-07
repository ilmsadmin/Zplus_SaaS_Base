# Redis Key Strategies for Multi-tenant Architecture
# Prefix-per-tenant approach to ensure data isolation

# Key Naming Convention:
# Format: {tenant_slug}:{service}:{resource}:{identifier}
# Example: acme_corp:auth:session:user_123
#          startup_inc:cache:product:prod_456

# 1. Authentication & Session Management
# User sessions
# Key: {tenant_slug}:auth:session:{session_token}
# TTL: 24 hours (configurable)
# Value: JSON object with user data
SET acme_corp:auth:session:sess_abc123 '{"user_id":"user_123","email":"john@acme.com","role":"admin","expires_at":"2025-08-08T10:00:00Z"}' EX 86400

# Refresh tokens
# Key: {tenant_slug}:auth:refresh:{user_id}
# TTL: 30 days (configurable)
# Value: Refresh token string
SET acme_corp:auth:refresh:user_123 "refresh_token_xyz789" EX 2592000

# Password reset tokens
# Key: {tenant_slug}:auth:reset:{token}
# TTL: 1 hour
# Value: User ID
SET acme_corp:auth:reset:reset_token_abc "user_123" EX 3600

# Email verification tokens
# Key: {tenant_slug}:auth:verify:{token}
# TTL: 24 hours
# Value: User email
SET acme_corp:auth:verify:verify_token_def "john@acme.com" EX 86400

# Failed login attempts (rate limiting)
# Key: {tenant_slug}:auth:failed:{ip_or_email}
# TTL: 15 minutes
# Value: Attempt count
INCR acme_corp:auth:failed:192.168.1.100
EXPIRE acme_corp:auth:failed:192.168.1.100 900

# 2. Caching
# User profile cache
# Key: {tenant_slug}:cache:user:{user_id}
# TTL: 1 hour
# Value: JSON user profile
SET acme_corp:cache:user:user_123 '{"id":"user_123","name":"John Doe","email":"john@acme.com","role":"admin"}' EX 3600

# Product cache
# Key: {tenant_slug}:cache:product:{product_id}
# TTL: 30 minutes
# Value: JSON product data
SET acme_corp:cache:product:prod_456 '{"id":"prod_456","name":"Laptop","price":999.99,"stock":50}' EX 1800

# Categories cache
# Key: {tenant_slug}:cache:categories:all
# TTL: 2 hours
# Value: JSON array of categories
SET acme_corp:cache:categories:all '[{"id":"cat_1","name":"Electronics"},{"id":"cat_2","name":"Books"}]' EX 7200

# Search results cache
# Key: {tenant_slug}:cache:search:{hash_of_query}
# TTL: 10 minutes
# Value: JSON search results
SET acme_corp:cache:search:f1a2b3c4d5 '{"products":[...],"total":150,"page":1}' EX 600

# API response cache
# Key: {tenant_slug}:cache:api:{endpoint_hash}:{params_hash}
# TTL: 5 minutes
# Value: JSON API response
SET acme_corp:cache:api:get_products:filters_abc '{"data":[...],"meta":{"total":100}}' EX 300

# 3. Rate Limiting
# API rate limiting per user
# Key: {tenant_slug}:rate:api:{user_id}:{endpoint}
# TTL: 1 minute (sliding window)
# Value: Request count
INCR acme_corp:rate:api:user_123:get_products
EXPIRE acme_corp:rate:api:user_123:get_products 60

# Global rate limiting per tenant
# Key: {tenant_slug}:rate:global:{endpoint}
# TTL: 1 minute
# Value: Request count
INCR acme_corp:rate:global:get_products
EXPIRE acme_corp:rate:global:get_products 60

# IP-based rate limiting
# Key: {tenant_slug}:rate:ip:{ip_address}
# TTL: 1 hour
# Value: Request count
INCR acme_corp:rate:ip:192.168.1.100
EXPIRE acme_corp:rate:ip:192.168.1.100 3600

# 4. Real-time Features
# WebSocket connections
# Key: {tenant_slug}:ws:connections:{user_id}
# TTL: No expiry (removed on disconnect)
# Value: Set of connection IDs
SADD acme_corp:ws:connections:user_123 "conn_abc123"

# Online users
# Key: {tenant_slug}:online:users
# TTL: 5 minutes (refreshed on activity)
# Value: Set of user IDs
SADD acme_corp:online:users "user_123"
EXPIRE acme_corp:online:users 300

# Real-time notifications
# Key: {tenant_slug}:notifications:{user_id}
# TTL: 24 hours
# Value: List of notifications (JSON strings)
LPUSH acme_corp:notifications:user_123 '{"id":"notif_1","title":"New Order","message":"Order #123 received"}'
EXPIRE acme_corp:notifications:user_123 86400

# 5. Shopping Cart (for POS module)
# User shopping cart
# Key: {tenant_slug}:cart:{user_id}
# TTL: 7 days
# Value: JSON cart data
SET acme_corp:cart:user_123 '{"items":[{"product_id":"prod_456","quantity":2,"price":999.99}],"total":1999.98}' EX 604800

# Temporary cart for anonymous users
# Key: {tenant_slug}:cart:temp:{session_id}
# TTL: 1 day
# Value: JSON cart data
SET acme_corp:cart:temp:sess_xyz789 '{"items":[...],"total":299.99}' EX 86400

# 6. File Upload Progress
# File upload progress tracking
# Key: {tenant_slug}:upload:{upload_id}
# TTL: 1 hour
# Value: Progress percentage
SET acme_corp:upload:upload_abc123 "45" EX 3600

# 7. Background Jobs & Queues
# Job queues (using Redis lists)
# Key: {tenant_slug}:queue:{queue_name}
# Value: JSON job data
LPUSH acme_corp:queue:email '{"type":"welcome_email","user_id":"user_123","data":{"email":"john@acme.com"}}'
LPUSH acme_corp:queue:export '{"type":"product_export","user_id":"user_123","format":"csv"}'

# Job status tracking
# Key: {tenant_slug}:job:{job_id}
# TTL: 24 hours
# Value: JSON job status
SET acme_corp:job:job_abc123 '{"status":"processing","progress":75,"started_at":"2025-08-07T10:00:00Z"}' EX 86400

# 8. Analytics & Metrics
# Daily page views
# Key: {tenant_slug}:analytics:page_views:{date}
# TTL: 90 days
# Value: Hash of page -> view count
HINCRBY acme_corp:analytics:page_views:2025-08-07 "/dashboard" 1
EXPIRE acme_corp:analytics:page_views:2025-08-07 7776000

# Real-time metrics
# Key: {tenant_slug}:metrics:realtime
# TTL: 1 hour
# Value: JSON metrics data
SET acme_corp:metrics:realtime '{"active_users":45,"orders_today":23,"revenue_today":15000}' EX 3600

# 9. Feature Flags
# Tenant feature flags
# Key: {tenant_slug}:features:{feature_name}
# TTL: No expiry (manual management)
# Value: Boolean or JSON config
SET acme_corp:features:advanced_analytics "true"
SET acme_corp:features:custom_domain '{"enabled":true,"domain":"shop.acme.com"}'

# 10. System Health & Monitoring
# Tenant health status
# Key: {tenant_slug}:health:status
# TTL: 5 minutes
# Value: JSON health data
SET acme_corp:health:status '{"database":"healthy","storage":"healthy","apis":"healthy","last_check":"2025-08-07T10:30:00Z"}' EX 300

# Redis Commands for Management:

# Get all keys for a tenant (use with caution in production)
# KEYS acme_corp:*

# Monitor tenant activity
# MONITOR (filter by tenant prefix)

# Get tenant memory usage
# MEMORY USAGE acme_corp:cache:user:user_123

# Flush all data for a tenant (use with extreme caution)
# EVAL "return redis.call('del', unpack(redis.call('keys', ARGV[1])))" 0 acme_corp:*

# Example Redis configuration for multi-tenant setup:
# maxmemory 2gb
# maxmemory-policy allkeys-lru
# save 900 1
# save 300 10
# save 60 10000

print("Redis key strategies documentation created for multi-tenant architecture")
