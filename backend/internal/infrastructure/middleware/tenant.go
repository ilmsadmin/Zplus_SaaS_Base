package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// TenantContext holds tenant information in request context
type TenantContext struct {
	TenantID     uuid.UUID
	Subdomain    string
	CustomDomain string
	Tenant       *domain.Tenant
}

// TenantMiddleware extracts tenant information from request and validates access
func TenantMiddleware(tenantRepo domain.TenantRepository, domainRepo domain.TenantDomainRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tenantCtx TenantContext
		var tenant *domain.Tenant
		var err error

		host := c.Get("Host")
		if host == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Host header is required",
			})
		}

		// Remove port if present
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}

		// Check if it's admin access (no tenant required)
		if host == "admin.zplus.io" || host == "localhost" || strings.HasPrefix(host, "admin.") {
			// System admin access - no tenant context needed
			c.Locals("is_admin", true)
			c.Locals("tenant", nil)
			return c.Next()
		}

		// Check if it's a subdomain pattern (*.zplus.io)
		if strings.HasSuffix(host, ".zplus.io") {
			subdomain := strings.TrimSuffix(host, ".zplus.io")
			if subdomain != "" && subdomain != "admin" {
				tenantCtx.Subdomain = subdomain
				// Look up tenant by subdomain
				tenant, err = tenantRepo.GetBySubdomain(context.Background(), subdomain)
				if err != nil {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error":   "tenant_not_found",
						"message": fmt.Sprintf("Tenant with subdomain '%s' not found", subdomain),
					})
				}
			}
		} else if host != "zplus.io" {
			// It's a custom domain
			tenantCtx.CustomDomain = host
			// Look up tenant by custom domain
			domain, err := domainRepo.GetByDomain(context.Background(), host)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "domain_not_found",
					"message": fmt.Sprintf("Domain '%s' not found", host),
				})
			}

			if !domain.IsVerified {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "domain_not_verified",
					"message": "Domain is not verified",
				})
			}

			// Get tenant by domain
			tenant, err = tenantRepo.GetByID(context.Background(), domain.TenantID)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "tenant_not_found",
					"message": "Tenant not found for domain",
				})
			}
		}

		// Check tenant status
		if tenant != nil {
			if tenant.Status != domain.StatusActive {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "tenant_inactive",
					"message": "Tenant is not active",
				})
			}

			tenantCtx.TenantID = tenant.ID
			tenantCtx.Tenant = tenant
		}

		// Set tenant context in locals
		c.Locals("tenant", tenantCtx)
		c.Locals("tenant_id", tenantCtx.TenantID.String())
		c.Locals("is_admin", false)

		return c.Next()
	}
}

// GetTenantFromContext retrieves tenant context from fiber context
func GetTenantFromContext(c *fiber.Ctx) (TenantContext, bool) {
	tenantCtx, ok := c.Locals("tenant").(TenantContext)
	return tenantCtx, ok
}

// RequireTenant middleware ensures a tenant is present in the request
func RequireTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantCtx, ok := GetTenantFromContext(c)
		if !ok || (tenantCtx.Subdomain == "" && tenantCtx.CustomDomain == "") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tenant information is required",
			})
		}
		return c.Next()
	}
}
