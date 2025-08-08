package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/application/dtos"
	"github.com/ilmsadmin/zplus-saas-base/internal/application/services"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// ReportingAnalyticsHandler handles reporting and analytics API endpoints
type ReportingAnalyticsHandler struct {
	reportingService services.ReportingAnalyticsService
	logger           *zap.Logger
}

// NewReportingAnalyticsHandler creates a new reporting analytics handler
func NewReportingAnalyticsHandler(reportingService services.ReportingAnalyticsService, logger *zap.Logger) *ReportingAnalyticsHandler {
	return &ReportingAnalyticsHandler{
		reportingService: reportingService,
		logger:           logger,
	}
}

// ============================
// Analytics Reports Endpoints
// ============================

// CreateReport creates a new analytics report
// @Summary Create Analytics Report
// @Description Create a new analytics report for the tenant
// @Tags Analytics Reports
// @Accept json
// @Produce json
// @Param request body dtos.CreateAnalyticsReportRequest true "Report creation request"
// @Success 201 {object} dtos.AnalyticsReportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/reports [post]
func (h *ReportingAnalyticsHandler) CreateReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(uuid.UUID)

	var req dtos.CreateAnalyticsReportRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate request
	if req.ReportType == "" || req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report type and title are required",
		})
	}

	report, err := h.reportingService.CreateReport(c.Context(), tenantID, userID, &req)
	if err != nil {
		h.logger.Error("Failed to create report", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create report",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// GetReport retrieves a specific report
// @Summary Get Analytics Report
// @Description Get analytics report by ID
// @Tags Analytics Reports
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} dtos.AnalyticsReportResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/reports/{id} [get]
func (h *ReportingAnalyticsHandler) GetReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	reportIDStr := c.Params("id")

	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID format",
		})
	}

	report, err := h.reportingService.GetReport(c.Context(), tenantID, reportID)
	if err != nil {
		h.logger.Error("Failed to get report", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Report not found",
		})
	}

	return c.JSON(report)
}

// UpdateReport updates an existing report
// @Summary Update Analytics Report
// @Description Update analytics report by ID
// @Tags Analytics Reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param request body dtos.UpdateAnalyticsReportRequest true "Report update request"
// @Success 200 {object} dtos.AnalyticsReportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/reports/{id} [put]
func (h *ReportingAnalyticsHandler) UpdateReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	reportIDStr := c.Params("id")

	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID format",
		})
	}

	var req dtos.UpdateAnalyticsReportRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	report, err := h.reportingService.UpdateReport(c.Context(), tenantID, reportID, &req)
	if err != nil {
		h.logger.Error("Failed to update report", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update report",
		})
	}

	return c.JSON(report)
}

// DeleteReport deletes a report
// @Summary Delete Analytics Report
// @Description Delete analytics report by ID
// @Tags Analytics Reports
// @Param id path string true "Report ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/reports/{id} [delete]
func (h *ReportingAnalyticsHandler) DeleteReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	reportIDStr := c.Params("id")

	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID format",
		})
	}

	if err := h.reportingService.DeleteReport(c.Context(), tenantID, reportID); err != nil {
		h.logger.Error("Failed to delete report", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete report",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListReports lists reports with pagination and filtering
// @Summary List Analytics Reports
// @Description List analytics reports with pagination and filtering
// @Tags Analytics Reports
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param report_type query string false "Filter by report type"
// @Param status query string false "Filter by status"
// @Success 200 {object} dtos.ReportListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/reports [get]
func (h *ReportingAnalyticsHandler) ListReports(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	reportType := c.Query("report_type")
	status := c.Query("status")

	filter := &domain.ReportFilter{
		Page:  page,
		Limit: limit,
	}

	if reportType != "" {
		filter.ReportType = &reportType
	}
	if status != "" {
		filter.Status = &status
	}

	reports, err := h.reportingService.ListReports(c.Context(), tenantID, filter)
	if err != nil {
		h.logger.Error("Failed to list reports", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list reports",
		})
	}

	return c.JSON(reports)
}

// DownloadReport downloads a report file
// @Summary Download Analytics Report
// @Description Download analytics report file
// @Tags Analytics Reports
// @Param id path string true "Report ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/reports/{id}/download [get]
func (h *ReportingAnalyticsHandler) DownloadReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	reportIDStr := c.Params("id")

	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID format",
		})
	}

	fileURL, err := h.reportingService.DownloadReport(c.Context(), tenantID, reportID)
	if err != nil {
		h.logger.Error("Failed to get download URL", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Report not available for download",
		})
	}

	return c.JSON(fiber.Map{
		"download_url": fileURL,
	})
}

// GenerateReport manually triggers report generation
// @Summary Generate Analytics Report
// @Description Manually trigger report generation
// @Tags Analytics Reports
// @Param id path string true "Report ID"
// @Success 202 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/reports/{id}/generate [post]
func (h *ReportingAnalyticsHandler) GenerateReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	reportIDStr := c.Params("id")

	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID format",
		})
	}

	// Start generation asynchronously
	go func() {
		if err := h.reportingService.GenerateReport(c.Context(), tenantID, reportID); err != nil {
			h.logger.Error("Failed to generate report", zap.Error(err))
		}
	}()

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Report generation started",
	})
}

// ============================
// User Activity Endpoints
// ============================

// RecordUserActivity records user activity
// @Summary Record User Activity
// @Description Record user activity for analytics
// @Tags User Activity
// @Accept json
// @Produce json
// @Param request body dtos.RecordUserActivityRequest true "Activity recording request"
// @Success 201 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/analytics/activity [post]
func (h *ReportingAnalyticsHandler) RecordUserActivity(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	var req dtos.RecordUserActivityRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if err := h.reportingService.RecordUserActivity(c.Context(), tenantID, &req); err != nil {
		h.logger.Error("Failed to record user activity", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to record activity",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Activity recorded successfully",
	})
}

// GetUserActivityMetrics retrieves user activity metrics
// @Summary Get User Activity Metrics
// @Description Get user activity metrics with filtering
// @Tags User Activity
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param user_id query string false "Filter by user ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dtos.ActivityMetricsListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/analytics/activity [get]
func (h *ReportingAnalyticsHandler) GetUserActivityMetrics(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	userIDStr := c.Query("user_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	filter := &domain.ActivityMetricsFilter{
		Page:  page,
		Limit: limit,
	}

	if userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}

	if startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	metrics, err := h.reportingService.GetUserActivityMetrics(c.Context(), tenantID, filter)
	if err != nil {
		h.logger.Error("Failed to get activity metrics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get activity metrics",
		})
	}

	return c.JSON(metrics)
}

// GetUserActivitySummary retrieves user activity summary
// @Summary Get User Activity Summary
// @Description Get user activity summary for a specific user
// @Tags User Activity
// @Produce json
// @Param user_id path string true "User ID"
// @Param days query int false "Number of days" default(30)
// @Success 200 {object} dtos.UserActivitySummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/analytics/activity/users/{user_id}/summary [get]
func (h *ReportingAnalyticsHandler) GetUserActivitySummary(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userIDStr := c.Params("user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))

	summary, err := h.reportingService.GetUserActivitySummary(c.Context(), tenantID, userID, days)
	if err != nil {
		h.logger.Error("Failed to get user activity summary", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get activity summary",
		})
	}

	return c.JSON(summary)
}

// ============================
// System Metrics Endpoints
// ============================

// RecordSystemMetric records system usage metric
// @Summary Record System Metric
// @Description Record system usage metric
// @Tags System Metrics
// @Accept json
// @Produce json
// @Param request body dtos.RecordSystemMetricRequest true "Metric recording request"
// @Success 201 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/analytics/system/metrics [post]
func (h *ReportingAnalyticsHandler) RecordSystemMetric(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	var req dtos.RecordSystemMetricRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if err := h.reportingService.RecordSystemMetric(c.Context(), tenantID, &req); err != nil {
		h.logger.Error("Failed to record system metric", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to record metric",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Metric recorded successfully",
	})
}

// GetSystemMetrics retrieves system usage metrics
// @Summary Get System Metrics
// @Description Get system usage metrics with filtering
// @Tags System Metrics
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param metric_type query string false "Filter by metric type"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dtos.SystemMetricsListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/analytics/system/metrics [get]
func (h *ReportingAnalyticsHandler) GetSystemMetrics(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	metricType := c.Query("metric_type")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	filter := &domain.SystemMetricsFilter{
		Page:  page,
		Limit: limit,
	}

	if metricType != "" {
		filter.MetricType = &metricType
	}

	if startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	metrics, err := h.reportingService.GetSystemMetrics(c.Context(), tenantID, filter)
	if err != nil {
		h.logger.Error("Failed to get system metrics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get system metrics",
		})
	}

	return c.JSON(metrics)
}

// GetSystemOverview retrieves system overview
// @Summary Get System Overview
// @Description Get system overview dashboard data
// @Tags System Metrics
// @Produce json
// @Param days query int false "Number of days" default(7)
// @Success 200 {object} dtos.SystemOverviewResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/analytics/system/overview [get]
func (h *ReportingAnalyticsHandler) GetSystemOverview(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	days, _ := strconv.Atoi(c.Query("days", "7"))

	overview, err := h.reportingService.GetSystemOverview(c.Context(), tenantID, days)
	if err != nil {
		h.logger.Error("Failed to get system overview", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get system overview",
		})
	}

	return c.JSON(overview)
}

// ============================
// Dashboard Endpoints
// ============================

// GetDashboardStats retrieves dashboard statistics
// @Summary Get Dashboard Statistics
// @Description Get analytics dashboard statistics
// @Tags Dashboard
// @Produce json
// @Param period query string false "Period (week, month, quarter)" default("week")
// @Success 200 {object} dtos.DashboardStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/analytics/dashboard [get]
func (h *ReportingAnalyticsHandler) GetDashboardStats(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	period := c.Query("period", "week")

	stats, err := h.reportingService.GetDashboardStats(c.Context(), tenantID, period)
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get dashboard statistics",
		})
	}

	return c.JSON(stats)
}

// ============================
// Health Check
// ============================

// HealthCheck checks the health of the reporting service
// @Summary Health Check
// @Description Check the health of the reporting and analytics service
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/analytics/health [get]
func (h *ReportingAnalyticsHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"service":   "reporting-analytics",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
