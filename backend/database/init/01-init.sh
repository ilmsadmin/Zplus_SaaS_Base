#!/bin/bash

# PostgreSQL initialization script
# This script runs when PostgreSQL container starts for the first time

set -e

# Wait for PostgreSQL to be ready
until pg_isready -h localhost -p 5432 -U postgres; do
    echo "Waiting for PostgreSQL to be ready..."
    sleep 2
done

echo "PostgreSQL is ready. Running initialization..."

# Create additional database users
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create application user
    CREATE USER zplus_app WITH PASSWORD 'zplus_app_password_123';
    
    -- Grant permissions
    GRANT CONNECT ON DATABASE $POSTGRES_DB TO zplus_app;
    GRANT USAGE ON SCHEMA public TO zplus_app;
    GRANT CREATE ON SCHEMA public TO zplus_app;
    
    -- Create read-only user for reports/analytics
    CREATE USER zplus_readonly WITH PASSWORD 'zplus_readonly_password_123';
    GRANT CONNECT ON DATABASE $POSTGRES_DB TO zplus_readonly;
    GRANT USAGE ON SCHEMA public TO zplus_readonly;
    GRANT SELECT ON ALL TABLES IN SCHEMA public TO zplus_readonly;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO zplus_readonly;
    
    -- Create backup user
    CREATE USER zplus_backup WITH PASSWORD 'zplus_backup_password_123';
    GRANT CONNECT ON DATABASE $POSTGRES_DB TO zplus_backup;
    GRANT USAGE ON SCHEMA public TO zplus_backup;
    GRANT SELECT ON ALL TABLES IN SCHEMA public TO zplus_backup;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO zplus_backup;
    
    -- Enable required extensions
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
    CREATE EXTENSION IF NOT EXISTS "pg_trgm";
    CREATE EXTENSION IF NOT EXISTS "btree_gin";
    CREATE EXTENSION IF NOT EXISTS "btree_gist";
    
    -- Create helper functions
    CREATE OR REPLACE FUNCTION generate_api_key_prefix()
    RETURNS TEXT AS \$\$
    BEGIN
        RETURN 'zp_' || substring(encode(gen_random_bytes(8), 'base64'), 1, 8);
    END;
    \$\$ LANGUAGE plpgsql;
    
    CREATE OR REPLACE FUNCTION hash_api_key(key_text TEXT)
    RETURNS TEXT AS \$\$
    BEGIN
        RETURN encode(digest(key_text, 'sha256'), 'hex');
    END;
    \$\$ LANGUAGE plpgsql;
    
    -- Create function to get tenant from domain
    CREATE OR REPLACE FUNCTION get_tenant_from_domain(domain_name TEXT)
    RETURNS TEXT AS \$\$
    DECLARE
        tenant_result TEXT;
    BEGIN
        SELECT tenant_id INTO tenant_result 
        FROM tenant_domains 
        WHERE domain = domain_name AND verified = TRUE
        LIMIT 1;
        
        RETURN tenant_result;
    END;
    \$\$ LANGUAGE plpgsql;
EOSQL

echo "PostgreSQL initialization completed successfully!"
