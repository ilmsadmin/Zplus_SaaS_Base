package interfaces

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/application"
)

// SetupDomainRoutes sets up all domain management routes
func SetupDomainRoutes(app *fiber.App, domainHandler *DomainHandler, authMiddleware fiber.Handler, tenantMiddleware fiber.Handler, adminMiddleware fiber.Handler) {
	// API v1 routes
	api := app.Group("/api/v1")

	// Tenant domain management routes (requires authentication and tenant context)
	tenantDomains := api.Group("/tenants/:tenant_id/domains", authMiddleware, tenantMiddleware)
	{
		tenantDomains.Get("/", domainHandler.GetDomains)
		tenantDomains.Post("/", domainHandler.AddCustomDomain)
		tenantDomains.Post("/:domain_id/verify", domainHandler.VerifyDomain)
		tenantDomains.Get("/:domain_id/instructions", domainHandler.GetDomainValidationInstructions)
		tenantDomains.Delete("/:domain_id", domainHandler.DeleteDomain)
	}

	// Public domain status routes (no auth required for monitoring)
	domains := api.Group("/domains")
	{
		domains.Get("/:domain/status", domainHandler.GetDomainStatus)
		domains.Get("/:domain/metrics", domainHandler.GetDomainMetrics)
	}

	// Admin domain management routes (requires admin authentication)
	adminDomains := api.Group("/admin/domains", authMiddleware, adminMiddleware)
	{
		adminDomains.Get("/", domainHandler.ListActiveDomains)
	}
}

// DomainRoutesConfig contains the configuration for domain routes
type DomainRoutesConfig struct {
	DomainService    *application.DomainService
	Logger           *zap.Logger
	AuthMiddleware   fiber.Handler
	TenantMiddleware fiber.Handler
	AdminMiddleware  fiber.Handler
}

// SetupDomainRoutesWithConfig sets up domain routes with configuration
func SetupDomainRoutesWithConfig(app *fiber.App, config *DomainRoutesConfig) {
	domainHandler := NewDomainHandler(config.DomainService, config.Logger)
	SetupDomainRoutes(app, domainHandler, config.AuthMiddleware, config.TenantMiddleware, config.AdminMiddleware)
}

// Domain Management API Documentation
//
// This file sets up all the routes for domain management in the multi-tenant SaaS platform.
//
// ## Routes Overview:
//
// ### Tenant Domain Management (Authenticated):
// - GET    /api/v1/tenants/{tenant_id}/domains                    - List all domains for a tenant
// - POST   /api/v1/tenants/{tenant_id}/domains                    - Add a new custom domain
// - POST   /api/v1/tenants/{tenant_id}/domains/{id}/verify        - Verify domain ownership
// - GET    /api/v1/tenants/{tenant_id}/domains/{id}/instructions  - Get verification instructions
// - DELETE /api/v1/tenants/{tenant_id}/domains/{id}               - Remove a custom domain
//
// ### Public Domain Status (No Auth):
// - GET    /api/v1/domains/{domain}/status                        - Get domain status and health
// - GET    /api/v1/domains/{domain}/metrics                       - Get domain performance metrics
//
// ### Admin Domain Management (Admin Only):
// - GET    /api/v1/admin/domains                                  - List all domains across tenants
//
// ## Request/Response Examples:
//
// ### Add Custom Domain:
// ```
// POST /api/v1/tenants/acme/domains
// {
//   "domain": "app.acme.com",
//   "verification_method": "dns",
//   "auto_ssl": true,
//   "priority": 150
// }
//
// Response:
// {
//   "success": true,
//   "data": {
//     "domain": "app.acme.com",
//     "verification_token": "zplus-verify-abc123def456",
//     "verification_method": "dns",
//     "dns_record": {
//       "type": "TXT",
//       "name": "_zplus-verify.app.acme.com",
//       "value": "zplus-verify-abc123def456",
//       "ttl": 300
//     },
//     "instructions": "Add a TXT record to your DNS...",
//     "expires_at": "2025-08-08T10:00:00Z"
//   }
// }
// ```
//
// ### Verify Domain:
// ```
// POST /api/v1/tenants/acme/domains/550e8400-e29b-41d4-a716-446655440001/verify
//
// Response:
// {
//   "success": true,
//   "data": {
//     "domain": "app.acme.com",
//     "verified": true,
//     "ssl_enabled": false,
//     "status": "active",
//     "last_checked": "2025-08-07T15:30:00Z",
//     "health_status": "healthy"
//   }
// }
// ```
//
// ### Get Domain Status:
// ```
// GET /api/v1/domains/app.acme.com/status
//
// Response:
// {
//   "success": true,
//   "data": {
//     "domain": "app.acme.com",
//     "verified": true,
//     "ssl_enabled": true,
//     "status": "active",
//     "last_checked": "2025-08-07T15:30:00Z",
//     "ssl_expiry": "2025-11-05T10:00:00Z",
//     "health_status": "healthy",
//     "metrics": {
//       "uptime_percentage": 99.9,
//       "avg_response_time": 120,
//       "requests_per_hour": 1500
//     }
//   }
// }
// ```
//
// ### Get Domain Metrics:
// ```
// GET /api/v1/domains/app.acme.com/metrics?hours=24
//
// Response:
// {
//   "success": true,
//   "data": {
//     "domain": "app.acme.com",
//     "time_range": 24,
//     "metrics": {
//       "requests": {
//         "total": 12500,
//         "2xx": 11875,
//         "4xx": 500,
//         "5xx": 125,
//         "per_hour": [520, 480, 450, ...]
//       },
//       "response_time": {
//         "avg": 120,
//         "p50": 100,
//         "p95": 180,
//         "p99": 250
//       },
//       "ssl": {
//         "enabled": true,
//         "expires_at": "2025-11-05T10:00:00Z",
//         "auto_renew": true,
//         "issuer": "Let's Encrypt"
//       }
//     }
//   }
// }
// ```
//
// ## Middleware Requirements:
//
// ### Auth Middleware:
// - Validates JWT token
// - Sets user context
// - Required for all tenant and admin operations
//
// ### Tenant Middleware:
// - Validates tenant access
// - Ensures user belongs to the tenant
// - Sets tenant context
//
// ### Admin Middleware:
// - Validates admin permissions
// - Required for cross-tenant operations
// - Allows system-wide domain management
//
// ## Security Considerations:
//
// 1. **Domain Validation**: All domains are validated for format and ownership
// 2. **Tenant Isolation**: Users can only manage domains for their tenants
// 3. **Rate Limiting**: Domain operations are rate limited to prevent abuse
// 4. **DNS Verification**: Ownership is verified via DNS TXT records
// 5. **SSL Management**: Automatic SSL certificate provisioning and renewal
//
// ## Integration with Traefik:
//
// The domain routes integrate with our Traefik configuration to:
// - Dynamically update routing rules
// - Manage SSL certificates
// - Configure rate limiting per domain
// - Enable/disable health checks
// - Update domain routing cache for performance
