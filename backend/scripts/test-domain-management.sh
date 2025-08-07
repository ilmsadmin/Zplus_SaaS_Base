#!/bin/bash

# Domain Management API Testing Script
# Tests the API Gateway & Routing domain management functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="http://localhost:8080/api/v1"
TENANT_ID="acme"
ADMIN_TOKEN=""  # Would be obtained from login
TEST_DOMAIN="test.acme.com"

# Logging functions
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

# Test functions
test_api_health() {
    log_info "Testing API health..."
    
    response=$(curl -s -w "%{http_code}" -o /tmp/health_response.json "${API_BASE_URL}/health" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "API is healthy"
        cat /tmp/health_response.json | jq '.' 2>/dev/null || cat /tmp/health_response.json
    else
        log_error "API health check failed (HTTP $http_code)"
        return 1
    fi
}

test_get_domains() {
    log_info "Testing GET domains for tenant: $TENANT_ID"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/domains_response.json \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "${API_BASE_URL}/tenants/${TENANT_ID}/domains" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Retrieved domains successfully"
        cat /tmp/domains_response.json | jq '.' 2>/dev/null || cat /tmp/domains_response.json
    else
        log_warning "GET domains returned HTTP $http_code"
        cat /tmp/domains_response.json 2>/dev/null || echo "No response body"
    fi
}

test_add_custom_domain() {
    local domain="$1"
    log_info "Testing ADD custom domain: $domain"
    
    payload=$(cat <<EOF
{
    "domain": "$domain",
    "verification_method": "dns",
    "auto_ssl": true,
    "priority": 150
}
EOF
)
    
    response=$(curl -s -w "%{http_code}" -o /tmp/add_domain_response.json \
        -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "$payload" \
        "${API_BASE_URL}/tenants/${TENANT_ID}/domains" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "201" ]; then
        log_success "Added custom domain successfully"
        cat /tmp/add_domain_response.json | jq '.' 2>/dev/null || cat /tmp/add_domain_response.json
        
        # Extract domain ID for later tests
        DOMAIN_ID=$(cat /tmp/add_domain_response.json | jq -r '.data.domain_id // empty' 2>/dev/null || echo "")
        if [ -n "$DOMAIN_ID" ]; then
            echo "DOMAIN_ID=$DOMAIN_ID" > /tmp/test_vars.env
        fi
    else
        log_warning "ADD domain returned HTTP $http_code"
        cat /tmp/add_domain_response.json 2>/dev/null || echo "No response body"
    fi
}

test_get_domain_instructions() {
    local domain_id="$1"
    if [ -z "$domain_id" ]; then
        log_warning "No domain ID provided, skipping instructions test"
        return
    fi
    
    log_info "Testing GET domain verification instructions"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/instructions_response.json \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "${API_BASE_URL}/tenants/${TENANT_ID}/domains/${domain_id}/instructions" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Retrieved verification instructions"
        cat /tmp/instructions_response.json | jq '.' 2>/dev/null || cat /tmp/instructions_response.json
    else
        log_warning "GET instructions returned HTTP $http_code"
        cat /tmp/instructions_response.json 2>/dev/null || echo "No response body"
    fi
}

test_verify_domain() {
    local domain_id="$1"
    if [ -z "$domain_id" ]; then
        log_warning "No domain ID provided, skipping verification test"
        return
    fi
    
    log_info "Testing VERIFY domain"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/verify_response.json \
        -X POST \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "${API_BASE_URL}/tenants/${TENANT_ID}/domains/${domain_id}/verify" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Domain verification attempted"
        cat /tmp/verify_response.json | jq '.' 2>/dev/null || cat /tmp/verify_response.json
    else
        log_warning "VERIFY domain returned HTTP $http_code"
        cat /tmp/verify_response.json 2>/dev/null || echo "No response body"
    fi
}

test_domain_status() {
    local domain="$1"
    log_info "Testing GET domain status for: $domain"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/status_response.json \
        "${API_BASE_URL}/domains/${domain}/status" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Retrieved domain status"
        cat /tmp/status_response.json | jq '.' 2>/dev/null || cat /tmp/status_response.json
    else
        log_warning "GET domain status returned HTTP $http_code"
        cat /tmp/status_response.json 2>/dev/null || echo "No response body"
    fi
}

test_domain_metrics() {
    local domain="$1"
    log_info "Testing GET domain metrics for: $domain"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/metrics_response.json \
        "${API_BASE_URL}/domains/${domain}/metrics?hours=24" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Retrieved domain metrics"
        cat /tmp/metrics_response.json | jq '.' 2>/dev/null || cat /tmp/metrics_response.json
    else
        log_warning "GET domain metrics returned HTTP $http_code"
        cat /tmp/metrics_response.json 2>/dev/null || echo "No response body"
    fi
}

test_admin_list_domains() {
    log_info "Testing ADMIN list all domains"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/admin_domains_response.json \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "${API_BASE_URL}/admin/domains?limit=10&ssl_expiring=false" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Retrieved admin domains list"
        cat /tmp/admin_domains_response.json | jq '.' 2>/dev/null || cat /tmp/admin_domains_response.json
    else
        log_warning "ADMIN list domains returned HTTP $http_code"
        cat /tmp/admin_domains_response.json 2>/dev/null || echo "No response body"
    fi
}

test_delete_domain() {
    local domain_id="$1"
    if [ -z "$domain_id" ]; then
        log_warning "No domain ID provided, skipping deletion test"
        return
    fi
    
    log_info "Testing DELETE custom domain"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/delete_response.json \
        -X DELETE \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "${API_BASE_URL}/tenants/${TENANT_ID}/domains/${domain_id}" || true)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Deleted domain successfully"
        cat /tmp/delete_response.json | jq '.' 2>/dev/null || cat /tmp/delete_response.json
    else
        log_warning "DELETE domain returned HTTP $http_code"
        cat /tmp/delete_response.json 2>/dev/null || echo "No response body"
    fi
}

# DNS validation tests
test_dns_validation() {
    local domain="$1"
    log_info "Testing DNS validation for: $domain"
    
    # Test TXT record lookup
    log_info "Checking for verification TXT record..."
    verification_record="_zplus-verify.${domain}"
    
    if command -v dig >/dev/null 2>&1; then
        dig_result=$(dig +short TXT "$verification_record" 2>/dev/null || echo "")
        if [ -n "$dig_result" ]; then
            log_success "Found TXT record: $dig_result"
        else
            log_warning "No TXT record found for $verification_record"
        fi
    elif command -v nslookup >/dev/null 2>&1; then
        nslookup_result=$(nslookup -type=TXT "$verification_record" 2>/dev/null | grep -A1 "text =" | grep -v "text =" || echo "")
        if [ -n "$nslookup_result" ]; then
            log_success "Found TXT record: $nslookup_result"
        else
            log_warning "No TXT record found for $verification_record"
        fi
    else
        log_warning "Neither dig nor nslookup available for DNS testing"
    fi
    
    # Test domain reachability
    log_info "Testing domain reachability..."
    if command -v curl >/dev/null 2>&1; then
        if curl -s --connect-timeout 5 -I "http://$domain" >/dev/null 2>&1; then
            log_success "Domain $domain is reachable via HTTP"
        else
            log_warning "Domain $domain is not reachable via HTTP"
        fi
        
        if curl -s --connect-timeout 5 -I "https://$domain" >/dev/null 2>&1; then
            log_success "Domain $domain is reachable via HTTPS"
        else
            log_warning "Domain $domain is not reachable via HTTPS"
        fi
    fi
}

# Database validation tests
test_database_state() {
    log_info "Testing database state..."
    
    # Check if domain management tables exist
    if command -v docker >/dev/null 2>&1; then
        tables_exist=$(docker exec zplus_postgres psql -U postgres -d zplus_saas_base -t -c "
            SELECT COUNT(*) FROM information_schema.tables 
            WHERE table_name IN ('tenant_domains', 'domain_validation_logs', 'ssl_certificates', 'domain_routing_cache')
        " 2>/dev/null | tr -d ' ' || echo "0")
        
        if [ "$tables_exist" = "4" ]; then
            log_success "All domain management tables exist"
        else
            log_error "Missing domain management tables (found $tables_exist/4)"
        fi
        
        # Check for sample data
        domain_count=$(docker exec zplus_postgres psql -U postgres -d zplus_saas_base -t -c "SELECT COUNT(*) FROM tenant_domains" 2>/dev/null | tr -d ' ' || echo "0")
        log_info "Found $domain_count domains in database"
        
        cache_count=$(docker exec zplus_postgres psql -U postgres -d zplus_saas_base -t -c "SELECT COUNT(*) FROM domain_routing_cache" 2>/dev/null | tr -d ' ' || echo "0")
        log_info "Found $cache_count entries in routing cache"
    else
        log_warning "Docker not available, skipping database state check"
    fi
}

# Performance tests
test_api_performance() {
    log_info "Testing API performance..."
    
    if command -v curl >/dev/null 2>&1; then
        # Test response time for GET domains
        start_time=$(date +%s%N)
        curl -s "${API_BASE_URL}/tenants/${TENANT_ID}/domains" >/dev/null 2>&1 || true
        end_time=$(date +%s%N)
        response_time=$(( (end_time - start_time) / 1000000 ))
        
        if [ "$response_time" -lt 1000 ]; then
            log_success "API response time: ${response_time}ms (excellent)"
        elif [ "$response_time" -lt 2000 ]; then
            log_success "API response time: ${response_time}ms (good)"
        else
            log_warning "API response time: ${response_time}ms (slow)"
        fi
    fi
}

# Main test execution
main() {
    echo "=========================================="
    echo "Domain Management API Testing"
    echo "=========================================="
    echo ""
    
    log_info "Starting domain management API tests..."
    echo ""
    
    # Load test variables if they exist
    if [ -f /tmp/test_vars.env ]; then
        source /tmp/test_vars.env
    fi
    
    # Basic health and connectivity tests
    echo "--- Basic API Tests ---"
    test_api_health
    echo ""
    
    test_database_state
    echo ""
    
    test_api_performance
    echo ""
    
    # Domain management tests (will fail without auth, but shows API structure)
    echo "--- Domain Management Tests ---"
    test_get_domains
    echo ""
    
    test_add_custom_domain "$TEST_DOMAIN"
    echo ""
    
    if [ -n "$DOMAIN_ID" ]; then
        test_get_domain_instructions "$DOMAIN_ID"
        echo ""
        
        test_verify_domain "$DOMAIN_ID"
        echo ""
    fi
    
    # Public domain tests
    echo "--- Public Domain Tests ---"
    test_domain_status "demo.zplus.io"
    echo ""
    
    test_domain_metrics "demo.zplus.io"
    echo ""
    
    # Admin tests
    echo "--- Admin Tests ---"
    test_admin_list_domains
    echo ""
    
    # DNS validation tests
    echo "--- DNS Validation Tests ---"
    test_dns_validation "$TEST_DOMAIN"
    echo ""
    
    # Cleanup
    if [ -n "$DOMAIN_ID" ]; then
        echo "--- Cleanup ---"
        test_delete_domain "$DOMAIN_ID"
        echo ""
    fi
    
    log_success "Domain management API testing completed!"
    echo ""
    echo "Notes:"
    echo "- Authentication tests require valid JWT tokens"
    echo "- DNS validation requires actual DNS records"
    echo "- SSL tests require valid certificates"
    echo "- Some tests may fail in development environment"
    echo ""
    echo "Next steps:"
    echo "1. Set up authentication tokens"
    echo "2. Configure DNS records for test domains"
    echo "3. Deploy SSL certificates"
    echo "4. Test with real domain verification"
}

# Run tests
main "$@"
