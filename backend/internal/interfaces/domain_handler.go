package interfaces

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/application"
)

// DomainHandler handles domain management HTTP requests
type DomainHandler struct {
	domainService *application.DomainService
	logger        *zap.Logger
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(domainService *application.DomainService, logger *zap.Logger) *DomainHandler {
	return &DomainHandler{
		domainService: domainService,
		logger:        logger,
	}
}

// AddCustomDomain handles POST /api/v1/tenants/{tenant_id}/domains
func (h *DomainHandler) AddCustomDomain(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tenant_id is required",
		})
	}

	var req application.AddCustomDomainRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set tenant ID from URL parameter
	req.TenantID = tenantID

	// Validate request
	if req.Domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "domain is required",
		})
	}

	// Call domain service
	response, err := h.domainService.AddCustomDomain(c.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to add custom domain",
			zap.String("tenant_id", tenantID),
			zap.String("domain", req.Domain),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.logger.Info("Custom domain added successfully",
		zap.String("tenant_id", tenantID),
		zap.String("domain", req.Domain),
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetDomains handles GET /api/v1/tenants/{tenant_id}/domains
func (h *DomainHandler) GetDomains(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tenant_id is required",
		})
	}

	domains, err := h.domainService.GetDomainsByTenant(c.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get domains",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve domains",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    domains,
		"count":   len(domains),
	})
}

// VerifyDomain handles POST /api/v1/tenants/{tenant_id}/domains/{domain_id}/verify
func (h *DomainHandler) VerifyDomain(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	domainIDStr := c.Params("domain_id")

	if tenantID == "" || domainIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tenant_id and domain_id are required",
		})
	}

	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid domain_id format",
		})
	}

	status, err := h.domainService.VerifyDomain(c.Context(), domainID)
	if err != nil {
		h.logger.Error("Failed to verify domain",
			zap.String("tenant_id", tenantID),
			zap.String("domain_id", domainIDStr),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.logger.Info("Domain verification completed",
		zap.String("tenant_id", tenantID),
		zap.String("domain_id", domainIDStr),
		zap.Bool("verified", status.Verified),
	)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    status,
	})
}

// DeleteDomain handles DELETE /api/v1/tenants/{tenant_id}/domains/{domain_id}
func (h *DomainHandler) DeleteDomain(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	domainIDStr := c.Params("domain_id")

	if tenantID == "" || domainIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tenant_id and domain_id are required",
		})
	}

	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid domain_id format",
		})
	}

	err = h.domainService.DeleteCustomDomain(c.Context(), domainID, tenantID)
	if err != nil {
		h.logger.Error("Failed to delete domain",
			zap.String("tenant_id", tenantID),
			zap.String("domain_id", domainIDStr),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.logger.Info("Domain deleted successfully",
		zap.String("tenant_id", tenantID),
		zap.String("domain_id", domainIDStr),
	)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Domain deleted successfully",
	})
}

// GetDomainStatus handles GET /api/v1/domains/{domain}/status
func (h *DomainHandler) GetDomainStatus(c *fiber.Ctx) error {
	domain := c.Params("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "domain is required",
		})
	}

	// This would implement domain status checking
	// For now, return a placeholder response
	status := application.DomainStatus{
		Domain:       domain,
		Verified:     true,
		SSLEnabled:   true,
		Status:       "active",
		LastChecked:  time.Now(),
		HealthStatus: "healthy",
		Metrics: map[string]interface{}{
			"uptime_percentage": 99.9,
			"avg_response_time": 120,
			"requests_per_hour": 1500,
		},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    status,
	})
}

// GetDomainMetrics handles GET /api/v1/domains/{domain}/metrics
func (h *DomainHandler) GetDomainMetrics(c *fiber.Ctx) error {
	domain := c.Params("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "domain is required",
		})
	}

	// Parse query parameters
	hours := c.QueryInt("hours", 24)
	if hours > 168 { // Max 7 days
		hours = 168
	}

	// This would implement metrics collection
	// For now, return placeholder metrics
	metrics := map[string]interface{}{
		"domain":     domain,
		"time_range": hours,
		"metrics": map[string]interface{}{
			"requests": map[string]interface{}{
				"total":    12500,
				"2xx":      11875,
				"4xx":      500,
				"5xx":      125,
				"per_hour": []int{520, 480, 450, 600, 580, 620, 550, 490, 510, 530, 540, 560},
			},
			"response_time": map[string]interface{}{
				"avg":      120,
				"p50":      100,
				"p95":      180,
				"p99":      250,
				"per_hour": []int{115, 120, 125, 118, 122, 130, 135, 128, 124, 119, 121, 126},
			},
			"ssl": map[string]interface{}{
				"enabled":    true,
				"expires_at": time.Now().Add(60 * 24 * time.Hour).Format(time.RFC3339),
				"auto_renew": true,
				"issuer":     "Let's Encrypt",
			},
			"health_checks": map[string]interface{}{
				"total":        288,
				"successful":   286,
				"failed":       2,
				"success_rate": 99.3,
				"last_check":   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			},
		},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    metrics,
	})
}

// ListActiveDomains handles GET /api/v1/admin/domains (admin only)
func (h *DomainHandler) ListActiveDomains(c *fiber.Ctx) error {
	// Parse query parameters
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	status := c.Query("status", "")
	sslExpiring := c.QueryBool("ssl_expiring", false)

	if limit > 100 {
		limit = 100
	}

	// This would implement domain listing with filters
	// For now, return placeholder data
	domains := []map[string]interface{}{
		{
			"id":           "550e8400-e29b-41d4-a716-446655440001",
			"domain":       "admin.zplus.io",
			"tenant_id":    "system",
			"is_custom":    false,
			"verified":     true,
			"ssl_enabled":  true,
			"status":       "active",
			"ssl_expires":  time.Now().Add(60 * 24 * time.Hour).Format(time.RFC3339),
			"last_checked": time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
			"health":       "healthy",
		},
		{
			"id":           "550e8400-e29b-41d4-a716-446655440002",
			"domain":       "acme.zplus.io",
			"tenant_id":    "acme",
			"is_custom":    false,
			"verified":     true,
			"ssl_enabled":  true,
			"status":       "active",
			"ssl_expires":  time.Now().Add(45 * 24 * time.Hour).Format(time.RFC3339),
			"last_checked": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			"health":       "healthy",
		},
		{
			"id":           "550e8400-e29b-41d4-a716-446655440003",
			"domain":       "app.acme.com",
			"tenant_id":    "acme",
			"is_custom":    true,
			"verified":     false,
			"ssl_enabled":  false,
			"status":       "inactive",
			"ssl_expires":  nil,
			"last_checked": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"health":       "pending_verification",
		},
	}

	// Apply filters (placeholder logic)
	filteredDomains := domains
	if status != "" {
		var filtered []map[string]interface{}
		for _, domain := range domains {
			if domain["status"] == status {
				filtered = append(filtered, domain)
			}
		}
		filteredDomains = filtered
	}

	if sslExpiring {
		var filtered []map[string]interface{}
		for _, domain := range domains {
			if domain["ssl_expires"] != nil {
				// Logic to check if SSL is expiring soon
				filtered = append(filtered, domain)
			}
		}
		filteredDomains = filtered
	}

	// Apply pagination
	total := len(filteredDomains)
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedDomains := filteredDomains[start:end]

	return c.JSON(fiber.Map{
		"success": true,
		"data":    paginatedDomains,
		"meta": map[string]interface{}{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"count":  len(paginatedDomains),
		},
	})
}

// GetDomainValidationInstructions handles GET /api/v1/tenants/{tenant_id}/domains/{domain_id}/instructions
func (h *DomainHandler) GetDomainValidationInstructions(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	domainIDStr := c.Params("domain_id")

	if tenantID == "" || domainIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tenant_id and domain_id are required",
		})
	}

	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid domain_id format",
		})
	}

	// This would fetch the actual domain and return verification instructions
	// For now, return placeholder instructions
	instructions := map[string]interface{}{
		"domain_id":           domainID,
		"tenant_id":           tenantID,
		"verification_method": "dns",
		"dns_record": map[string]interface{}{
			"type":  "TXT",
			"name":  "_zplus-verify.app.acme.com",
			"value": "zplus-verify-abc123def456789",
			"ttl":   300,
		},
		"instructions": `To verify ownership of your domain, please add the following DNS record:

Record Type: TXT
Name: _zplus-verify.app.acme.com
Value: zplus-verify-abc123def456789
TTL: 300 (5 minutes)

Steps:
1. Log in to your DNS provider's control panel
2. Navigate to DNS management for your domain
3. Add a new TXT record with the above details
4. Wait for DNS propagation (usually 5-15 minutes)
5. Click "Verify Domain" button to complete verification

Note: The verification token expires in 24 hours. If you need a new token, delete and re-add the domain.`,
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"status":     "pending",
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    instructions,
	})
}
