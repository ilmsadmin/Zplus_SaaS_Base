package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/application/dtos"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// ReportingAnalyticsService provides reporting and analytics functionality
type ReportingAnalyticsService interface {
	// Analytics Reports
	CreateReport(ctx context.Context, tenantID string, userID uuid.UUID, req *dtos.CreateAnalyticsReportRequest) (*dtos.AnalyticsReportResponse, error)
	GetReport(ctx context.Context, tenantID string, reportID uuid.UUID) (*dtos.AnalyticsReportResponse, error)
	UpdateReport(ctx context.Context, tenantID string, reportID uuid.UUID, req *dtos.UpdateAnalyticsReportRequest) (*dtos.AnalyticsReportResponse, error)
	DeleteReport(ctx context.Context, tenantID string, reportID uuid.UUID) error
	ListReports(ctx context.Context, tenantID string, filter *domain.ReportFilter) (*dtos.ReportListResponse, error)
	GenerateReport(ctx context.Context, tenantID string, reportID uuid.UUID) error
	DownloadReport(ctx context.Context, tenantID string, reportID uuid.UUID) (string, error)

	// User Activity Analytics
	RecordUserActivity(ctx context.Context, tenantID string, req *dtos.RecordUserActivityRequest) error
	GetUserActivityMetrics(ctx context.Context, tenantID string, filter *domain.ActivityMetricsFilter) (*dtos.ActivityMetricsListResponse, error)
	GetUserActivitySummary(ctx context.Context, tenantID string, userID uuid.UUID, days int) (*dtos.UserActivitySummaryResponse, error)
	GetActivityTrends(ctx context.Context, tenantID string, startDate, endDate time.Time, groupBy string) ([]map[string]interface{}, error)

	// System Usage Analytics
	RecordSystemMetric(ctx context.Context, tenantID string, req *dtos.RecordSystemMetricRequest) error
	GetSystemMetrics(ctx context.Context, tenantID string, filter *domain.SystemMetricsFilter) (*dtos.SystemMetricsListResponse, error)
	GetSystemOverview(ctx context.Context, tenantID string, days int) (*dtos.SystemOverviewResponse, error)
	GetSystemStats(ctx context.Context, tenantID string, metricType string, startDate, endDate time.Time) (map[string]interface{}, error)

	// Report Exports
	CreateExport(ctx context.Context, tenantID string, userID uuid.UUID, req *dtos.CreateReportExportRequest) (*dtos.ReportExportResponse, error)
	GetExport(ctx context.Context, tenantID string, exportID uuid.UUID) (*dtos.ReportExportResponse, error)
	ListExports(ctx context.Context, tenantID string, userID *uuid.UUID, page, limit int) (*dtos.ExportListResponse, error)
	DownloadExport(ctx context.Context, tenantID string, exportID uuid.UUID) (string, error)

	// Report Schedules
	CreateSchedule(ctx context.Context, tenantID string, userID uuid.UUID, req *dtos.CreateReportScheduleRequest) (*dtos.ReportScheduleResponse, error)
	GetSchedule(ctx context.Context, tenantID string, scheduleID uuid.UUID) (*dtos.ReportScheduleResponse, error)
	UpdateSchedule(ctx context.Context, tenantID string, scheduleID uuid.UUID, req *dtos.UpdateReportScheduleRequest) (*dtos.ReportScheduleResponse, error)
	DeleteSchedule(ctx context.Context, tenantID string, scheduleID uuid.UUID) error
	ListSchedules(ctx context.Context, tenantID string, userID *uuid.UUID, page, limit int) (*dtos.ScheduleListResponse, error)
	ProcessScheduledReports(ctx context.Context) error

	// Dashboard
	GetDashboardStats(ctx context.Context, tenantID string, period string) (*dtos.DashboardStatsResponse, error)

	// Sales Reports (POS Integration)
	GenerateSalesReport(ctx context.Context, tenantID string, userID uuid.UUID, reportType string, startDate, endDate time.Time) (*dtos.AnalyticsReportResponse, error)
	GetSalesStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error)

	// Cleanup
	CleanupExpiredReports(ctx context.Context) error
	CleanupOldMetrics(ctx context.Context, retentionDays int) error
}

// ReportingAnalyticsServiceImpl implements ReportingAnalyticsService
type ReportingAnalyticsServiceImpl struct {
	reportRepo        domain.AnalyticsReportRepository
	activityRepo      domain.UserActivityMetricsRepository
	systemMetricsRepo domain.SystemUsageMetricsRepository
	exportRepo        domain.ReportExportRepository
	scheduleRepo      domain.ReportScheduleRepository
	userRepo          domain.UserRepository
	orderRepo         domain.OrderRepository
	logger            *zap.Logger
}

// NewReportingAnalyticsService creates a new reporting analytics service
func NewReportingAnalyticsService(
	reportRepo domain.AnalyticsReportRepository,
	activityRepo domain.UserActivityMetricsRepository,
	systemMetricsRepo domain.SystemUsageMetricsRepository,
	exportRepo domain.ReportExportRepository,
	scheduleRepo domain.ReportScheduleRepository,
	userRepo domain.UserRepository,
	orderRepo domain.OrderRepository,
	logger *zap.Logger,
) ReportingAnalyticsService {
	return &ReportingAnalyticsServiceImpl{
		reportRepo:        reportRepo,
		activityRepo:      activityRepo,
		systemMetricsRepo: systemMetricsRepo,
		exportRepo:        exportRepo,
		scheduleRepo:      scheduleRepo,
		userRepo:          userRepo,
		orderRepo:         orderRepo,
		logger:            logger,
	}
}

// ============================
// Analytics Reports
// ============================

// CreateReport creates a new analytics report
func (s *ReportingAnalyticsServiceImpl) CreateReport(ctx context.Context, tenantID string, userID uuid.UUID, req *dtos.CreateAnalyticsReportRequest) (*dtos.AnalyticsReportResponse, error) {
	s.logger.Info("Creating analytics report",
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID.String()),
		zap.String("report_type", req.ReportType),
	)

	// Validate user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create report entity
	report := &domain.AnalyticsReport{
		TenantID:       tenantID,
		UserID:         userID,
		ReportType:     req.ReportType,
		ReportSubtype:  req.ReportSubtype,
		Title:          req.Title,
		Description:    req.Description,
		PeriodStart:    req.PeriodStart,
		PeriodEnd:      req.PeriodEnd,
		Parameters:     req.Parameters,
		FileFormat:     req.FileFormat,
		ScheduledFor:   req.ScheduledFor,
		IsRecurring:    req.IsRecurring,
		RecurrenceRule: req.RecurrenceRule,
		Tags:           req.Tags,
		Metadata:       req.Metadata,
		Status:         "pending",
	}

	// Set default file format
	if report.FileFormat == "" {
		report.FileFormat = "json"
	}

	// Set expiration (30 days default)
	expiresAt := time.Now().AddDate(0, 0, 30)
	report.ExpiresAt = &expiresAt

	// Set next run if recurring
	if report.IsRecurring && report.RecurrenceRule != "" {
		nextRun := s.calculateNextRun(report.RecurrenceRule)
		report.NextRunAt = &nextRun
	}

	// Save to repository
	if err := s.reportRepo.Create(ctx, report); err != nil {
		s.logger.Error("Failed to create report", zap.Error(err))
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Start report generation if not scheduled
	if req.ScheduledFor == nil {
		go func() {
			if err := s.GenerateReport(context.Background(), tenantID, report.ID); err != nil {
				s.logger.Error("Failed to generate report", zap.Error(err))
			}
		}()
	}

	return s.toReportResponse(report, user), nil
}

// GetReport retrieves a report by ID
func (s *ReportingAnalyticsServiceImpl) GetReport(ctx context.Context, tenantID string, reportID uuid.UUID) (*dtos.AnalyticsReportResponse, error) {
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("report not found: %w", err)
	}

	if report.TenantID != tenantID {
		return nil, fmt.Errorf("report not found in tenant")
	}

	user, err := s.userRepo.GetByID(ctx, report.UserID)
	if err != nil {
		s.logger.Warn("User not found for report", zap.String("user_id", report.UserID.String()))
	}

	return s.toReportResponse(report, user), nil
}

// UpdateReport updates an existing report
func (s *ReportingAnalyticsServiceImpl) UpdateReport(ctx context.Context, tenantID string, reportID uuid.UUID, req *dtos.UpdateAnalyticsReportRequest) (*dtos.AnalyticsReportResponse, error) {
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("report not found: %w", err)
	}

	if report.TenantID != tenantID {
		return nil, fmt.Errorf("report not found in tenant")
	}

	// Update fields
	if req.Title != nil {
		report.Title = *req.Title
	}
	if req.Description != nil {
		report.Description = *req.Description
	}
	if req.Parameters != nil {
		report.Parameters = req.Parameters
	}
	if req.FileFormat != nil {
		report.FileFormat = *req.FileFormat
	}
	if req.IsRecurring != nil {
		report.IsRecurring = *req.IsRecurring
	}
	if req.RecurrenceRule != nil {
		report.RecurrenceRule = *req.RecurrenceRule
	}
	if req.Tags != nil {
		report.Tags = req.Tags
	}
	if req.Metadata != nil {
		report.Metadata = req.Metadata
	}

	// Update next run if recurring changed
	if req.IsRecurring != nil && *req.IsRecurring && req.RecurrenceRule != nil {
		nextRun := s.calculateNextRun(*req.RecurrenceRule)
		report.NextRunAt = &nextRun
	}

	if err := s.reportRepo.Update(ctx, report); err != nil {
		s.logger.Error("Failed to update report", zap.Error(err))
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, report.UserID)
	if err != nil {
		s.logger.Warn("User not found for report", zap.String("user_id", report.UserID.String()))
	}

	return s.toReportResponse(report, user), nil
}

// DeleteReport deletes a report
func (s *ReportingAnalyticsServiceImpl) DeleteReport(ctx context.Context, tenantID string, reportID uuid.UUID) error {
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("report not found: %w", err)
	}

	if report.TenantID != tenantID {
		return fmt.Errorf("report not found in tenant")
	}

	if err := s.reportRepo.Delete(ctx, reportID); err != nil {
		s.logger.Error("Failed to delete report", zap.Error(err))
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}

// ListReports lists reports with pagination and filtering
func (s *ReportingAnalyticsServiceImpl) ListReports(ctx context.Context, tenantID string, filter *domain.ReportFilter) (*dtos.ReportListResponse, error) {
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	reports, total, err := s.reportRepo.GetByTenantID(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error("Failed to list reports", zap.Error(err))
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	responses := make([]*dtos.AnalyticsReportResponse, len(reports))
	for i, report := range reports {
		user, err := s.userRepo.GetByID(ctx, report.UserID)
		if err != nil {
			s.logger.Warn("User not found for report", zap.String("user_id", report.UserID.String()))
		}
		responses[i] = s.toReportResponse(report, user)
	}

	return &dtos.ReportListResponse{
		Reports:    responses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit))),
	}, nil
}

// GenerateReport generates report data and files
func (s *ReportingAnalyticsServiceImpl) GenerateReport(ctx context.Context, tenantID string, reportID uuid.UUID) error {
	s.logger.Info("Generating report", zap.String("report_id", reportID.String()))

	// Update status to processing
	if err := s.reportRepo.UpdateStatus(ctx, reportID, "processing", ""); err != nil {
		return fmt.Errorf("failed to update report status: %w", err)
	}

	// Get report details
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("report not found: %w", err)
	}

	// Generate report data based on type
	var reportData map[string]interface{}
	var summary map[string]interface{}

	switch report.ReportType {
	case "sales":
		reportData, summary, err = s.generateSalesReportData(ctx, tenantID, report.PeriodStart, report.PeriodEnd, report.Parameters)
	case "users":
		reportData, summary, err = s.generateUserReportData(ctx, tenantID, report.PeriodStart, report.PeriodEnd, report.Parameters)
	case "system":
		reportData, summary, err = s.generateSystemReportData(ctx, tenantID, report.PeriodStart, report.PeriodEnd, report.Parameters)
	default:
		err = fmt.Errorf("unsupported report type: %s", report.ReportType)
	}

	if err != nil {
		s.logger.Error("Failed to generate report data", zap.Error(err))
		if updateErr := s.reportRepo.MarkFailed(ctx, reportID, err.Error()); updateErr != nil {
			s.logger.Error("Failed to mark report as failed", zap.Error(updateErr))
		}
		return fmt.Errorf("failed to generate report data: %w", err)
	}

	// Update report with data
	report.ReportData = reportData
	report.Summary = summary
	if err := s.reportRepo.Update(ctx, report); err != nil {
		s.logger.Error("Failed to update report data", zap.Error(err))
		return fmt.Errorf("failed to update report data: %w", err)
	}

	// Generate file if needed
	filePath := ""
	fileURL := ""
	fileSize := int64(0)

	if report.FileFormat != "json" {
		filePath, fileURL, fileSize, err = s.generateReportFile(ctx, report, reportData)
		if err != nil {
			s.logger.Error("Failed to generate report file", zap.Error(err))
			if updateErr := s.reportRepo.MarkFailed(ctx, reportID, err.Error()); updateErr != nil {
				s.logger.Error("Failed to mark report as failed", zap.Error(updateErr))
			}
			return fmt.Errorf("failed to generate report file: %w", err)
		}
	}

	// Mark as completed
	if err := s.reportRepo.MarkCompleted(ctx, reportID, filePath, fileURL, fileSize); err != nil {
		s.logger.Error("Failed to mark report as completed", zap.Error(err))
		return fmt.Errorf("failed to mark report as completed: %w", err)
	}

	s.logger.Info("Report generated successfully", zap.String("report_id", reportID.String()))
	return nil
}

// DownloadReport tracks download and returns file URL
func (s *ReportingAnalyticsServiceImpl) DownloadReport(ctx context.Context, tenantID string, reportID uuid.UUID) (string, error) {
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return "", fmt.Errorf("report not found: %w", err)
	}

	if report.TenantID != tenantID {
		return "", fmt.Errorf("report not found in tenant")
	}

	if report.Status != "completed" {
		return "", fmt.Errorf("report is not ready for download: status=%s", report.Status)
	}

	// Update download tracking
	if err := s.reportRepo.IncrementDownloadCount(ctx, reportID); err != nil {
		s.logger.Warn("Failed to track download count", zap.Error(err))
	}

	if err := s.reportRepo.UpdateLastDownload(ctx, reportID); err != nil {
		s.logger.Warn("Failed to update last download time", zap.Error(err))
	}

	if report.FileURL != "" {
		return report.FileURL, nil
	}

	// Return JSON data URL for direct download
	return fmt.Sprintf("/api/reports/%s/data", reportID.String()), nil
}

// ============================
// Helper Methods
// ============================

// toReportResponse converts domain model to DTO
func (s *ReportingAnalyticsServiceImpl) toReportResponse(report *domain.AnalyticsReport, user *domain.User) *dtos.AnalyticsReportResponse {
	return &dtos.AnalyticsReportResponse{
		ID:              report.ID,
		ReportType:      report.ReportType,
		ReportSubtype:   report.ReportSubtype,
		Title:           report.Title,
		Description:     report.Description,
		PeriodStart:     report.PeriodStart,
		PeriodEnd:       report.PeriodEnd,
		Parameters:      report.Parameters,
		Summary:         report.Summary,
		FileURL:         report.FileURL,
		FileFormat:      report.FileFormat,
		FileSize:        report.FileSize,
		Status:          report.Status,
		ErrorMessage:    report.ErrorMessage,
		ScheduledFor:    report.ScheduledFor,
		ProcessingStats: report.ProcessingStats,
		ExpiresAt:       report.ExpiresAt,
		DownloadCount:   report.DownloadCount,
		LastDownloadAt:  report.LastDownloadAt,
		IsRecurring:     report.IsRecurring,
		RecurrenceRule:  report.RecurrenceRule,
		NextRunAt:       report.NextRunAt,
		Tags:            report.Tags,
		CreatedAt:       report.CreatedAt,
		UpdatedAt:       report.UpdatedAt,
		CompletedAt:     report.CompletedAt,
	}
}

// calculateNextRun calculates next run time for recurring reports
func (s *ReportingAnalyticsServiceImpl) calculateNextRun(recurrenceRule string) time.Time {
	// For now, implement basic rules. In production, use a proper cron parser
	switch recurrenceRule {
	case "daily":
		return time.Now().AddDate(0, 0, 1)
	case "weekly":
		return time.Now().AddDate(0, 0, 7)
	case "monthly":
		return time.Now().AddDate(0, 1, 0)
	default:
		return time.Now().AddDate(0, 0, 1) // Default to daily
	}
}

// generateSalesReportData generates sales report data
func (s *ReportingAnalyticsServiceImpl) generateSalesReportData(ctx context.Context, tenantID string, startDate, endDate time.Time, parameters map[string]interface{}) (map[string]interface{}, map[string]interface{}, error) {
	// This would integrate with POS module
	// For now, return sample data structure
	reportData := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"sales": map[string]interface{}{
			"total_revenue": 0.0,
			"total_orders":  0,
			"avg_order":     0.0,
		},
		"top_products": []map[string]interface{}{},
		"daily_sales":  []map[string]interface{}{},
	}

	summary := map[string]interface{}{
		"total_revenue":  0.0,
		"total_orders":   0,
		"growth_percent": 0.0,
		"top_product":    "",
		"generated_at":   time.Now(),
	}

	return reportData, summary, nil
}

// generateUserReportData generates user activity report data
func (s *ReportingAnalyticsServiceImpl) generateUserReportData(ctx context.Context, tenantID string, startDate, endDate time.Time, parameters map[string]interface{}) (map[string]interface{}, map[string]interface{}, error) {
	// Get user activity metrics
	filter := &domain.ActivityMetricsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     1000,
		Page:      1,
	}

	metrics, _, err := s.activityRepo.GetByTenantAndDateRange(ctx, tenantID, startDate, endDate, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user metrics: %w", err)
	}

	// Process metrics
	totalPageViews := 0
	totalSessions := 0
	uniqueUsers := make(map[uuid.UUID]bool)
	deviceTypes := make(map[string]int)

	for _, metric := range metrics {
		totalPageViews += metric.PageViews
		totalSessions += 1
		uniqueUsers[metric.UserID] = true
		if metric.DeviceType != "" {
			deviceTypes[metric.DeviceType]++
		}
	}

	reportData := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"overview": map[string]interface{}{
			"total_users":      len(uniqueUsers),
			"total_sessions":   totalSessions,
			"total_page_views": totalPageViews,
			"avg_session_time": 0, // Would calculate from session durations
		},
		"device_breakdown": deviceTypes,
		"daily_activity":   []map[string]interface{}{}, // Would group by date
		"top_pages":        []map[string]interface{}{}, // Would analyze page views
	}

	summary := map[string]interface{}{
		"active_users":    len(uniqueUsers),
		"total_sessions":  totalSessions,
		"engagement_rate": float64(totalPageViews) / float64(totalSessions),
		"generated_at":    time.Now(),
	}

	return reportData, summary, nil
}

// generateSystemReportData generates system usage report data
func (s *ReportingAnalyticsServiceImpl) generateSystemReportData(ctx context.Context, tenantID string, startDate, endDate time.Time, parameters map[string]interface{}) (map[string]interface{}, map[string]interface{}, error) {
	// Get system metrics
	filter := &domain.SystemMetricsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     1000,
		Page:      1,
	}

	metrics, _, err := s.systemMetricsRepo.GetByTenantAndDateRange(ctx, tenantID, startDate, endDate, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	// Process metrics
	totalAPIsCalls := 0
	totalErrors := 0
	totalStorage := int64(0)
	var avgResponseTime float64

	for _, metric := range metrics {
		totalAPIsCalls += metric.APICallsTotal
		totalErrors += metric.APICallsError
		totalStorage += metric.StorageUsed
		avgResponseTime += float64(metric.APIResponseTime)
	}

	if len(metrics) > 0 {
		avgResponseTime = avgResponseTime / float64(len(metrics))
	}

	errorRate := float64(0)
	if totalAPIsCalls > 0 {
		errorRate = float64(totalErrors) / float64(totalAPIsCalls) * 100
	}

	reportData := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"api_usage": map[string]interface{}{
			"total_calls":       totalAPIsCalls,
			"total_errors":      totalErrors,
			"error_rate":        errorRate,
			"avg_response_time": avgResponseTime,
		},
		"storage_usage": map[string]interface{}{
			"total_storage": totalStorage,
			"storage_trend": []map[string]interface{}{}, // Would calculate trend
		},
		"performance": map[string]interface{}{
			"uptime_percent":    99.9, // Would calculate from metrics
			"avg_response_time": avgResponseTime,
		},
	}

	summary := map[string]interface{}{
		"api_calls":       totalAPIsCalls,
		"error_rate":      errorRate,
		"storage_used":    totalStorage,
		"avg_response_ms": avgResponseTime,
		"generated_at":    time.Now(),
	}

	return reportData, summary, nil
}

// generateReportFile generates report file in specified format
func (s *ReportingAnalyticsServiceImpl) generateReportFile(ctx context.Context, report *domain.AnalyticsReport, data map[string]interface{}) (string, string, int64, error) {
	// This would integrate with file management service
	// For now, return placeholder values

	fileName := fmt.Sprintf("report_%s_%s.%s", report.ReportType, report.ID.String()[:8], report.FileFormat)
	filePath := fmt.Sprintf("/reports/%s/%s", report.TenantID, fileName)
	fileURL := fmt.Sprintf("/api/files/reports/%s", fileName)
	fileSize := int64(1024) // Placeholder size

	s.logger.Info("Report file generated",
		zap.String("file_path", filePath),
		zap.String("format", report.FileFormat),
	)

	return filePath, fileURL, fileSize, nil
}

// ============================
// User Activity Analytics
// ============================

// RecordUserActivity records user activity metrics
func (s *ReportingAnalyticsServiceImpl) RecordUserActivity(ctx context.Context, tenantID string, req *dtos.RecordUserActivityRequest) error {
	s.logger.Debug("Recording user activity",
		zap.String("tenant_id", tenantID),
		zap.String("user_id", req.UserID.String()),
		zap.String("session_id", req.SessionID),
	)

	// Create or update activity metrics for today
	activityData := map[string]interface{}{
		"page":       req.Page,
		"action":     req.Action,
		"ip_address": req.IPAddress,
		"user_agent": req.Browser + " on " + req.Platform,
	}

	if req.ActivityData != nil {
		for k, v := range req.ActivityData {
			activityData[k] = v
		}
	}

	return s.activityRepo.RecordUserActivity(ctx, tenantID, req.UserID, req.SessionID, activityData)
}

// GetUserActivityMetrics retrieves user activity metrics with filtering
func (s *ReportingAnalyticsServiceImpl) GetUserActivityMetrics(ctx context.Context, tenantID string, filter *domain.ActivityMetricsFilter) (*dtos.ActivityMetricsListResponse, error) {
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	startDate := time.Now().AddDate(0, 0, -30) // Default to last 30 days
	endDate := time.Now()

	if filter.StartDate != nil {
		startDate = *filter.StartDate
	}
	if filter.EndDate != nil {
		endDate = *filter.EndDate
	}

	metrics, total, err := s.activityRepo.GetByTenantAndDateRange(ctx, tenantID, startDate, endDate, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity metrics: %w", err)
	}

	responses := make([]*dtos.UserActivityMetricsResponse, len(metrics))
	for i, metric := range metrics {
		responses[i] = s.toActivityMetricsResponse(metric)
	}

	return &dtos.ActivityMetricsListResponse{
		Metrics:    responses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit))),
	}, nil
}

// GetUserActivitySummary retrieves user activity summary
func (s *ReportingAnalyticsServiceImpl) GetUserActivitySummary(ctx context.Context, tenantID string, userID uuid.UUID, days int) (*dtos.UserActivitySummaryResponse, error) {
	summary, err := s.activityRepo.GetUserSummary(ctx, tenantID, userID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get user summary: %w", err)
	}

	// Convert to response DTO
	response := &dtos.UserActivitySummaryResponse{
		UserID: userID,
	}

	if totalPageViews, ok := summary["total_page_views"].(int); ok {
		response.TotalPageViews = totalPageViews
	}
	if totalSessions, ok := summary["total_sessions"].(int); ok {
		response.TotalSessions = totalSessions
	}
	if totalDuration, ok := summary["total_duration"].(int); ok {
		response.TotalDuration = totalDuration
	}
	if avgSessionTime, ok := summary["avg_session_time"].(int); ok {
		response.AvgSessionTime = avgSessionTime
	}
	if totalActions, ok := summary["total_actions"].(int); ok {
		response.TotalActions = totalActions
	}
	if loginCount, ok := summary["login_count"].(int); ok {
		response.LoginCount = loginCount
	}

	// Set defaults for complex fields
	response.TopPages = []map[string]interface{}{}
	response.DeviceBreakdown = make(map[string]int)
	response.DailyActivities = []dtos.UserActivityMetricsResponse{}

	return response, nil
}

// GetActivityTrends retrieves activity trends
func (s *ReportingAnalyticsServiceImpl) GetActivityTrends(ctx context.Context, tenantID string, startDate, endDate time.Time, groupBy string) ([]map[string]interface{}, error) {
	return s.activityRepo.GetActivityTrends(ctx, tenantID, startDate, endDate, groupBy)
}

// ============================
// System Usage Analytics
// ============================

// RecordSystemMetric records system usage metrics
func (s *ReportingAnalyticsServiceImpl) RecordSystemMetric(ctx context.Context, tenantID string, req *dtos.RecordSystemMetricRequest) error {
	s.logger.Debug("Recording system metric",
		zap.String("tenant_id", tenantID),
		zap.String("metric_type", req.MetricType),
		zap.String("metric_name", req.MetricName),
	)

	date := time.Now().Truncate(24 * time.Hour)
	hour := time.Now().Hour()

	if req.Hour != nil {
		hour = *req.Hour
	}

	unit := req.MetricUnit
	if unit == "" {
		unit = "count"
	}

	return s.systemMetricsRepo.RecordCustomMetric(ctx, tenantID, date, hour, req.MetricType, req.MetricName, req.MetricValue, unit, req.CustomMetrics)
}

// GetSystemMetrics retrieves system usage metrics
func (s *ReportingAnalyticsServiceImpl) GetSystemMetrics(ctx context.Context, tenantID string, filter *domain.SystemMetricsFilter) (*dtos.SystemMetricsListResponse, error) {
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	startDate := time.Now().AddDate(0, 0, -7) // Default to last 7 days
	endDate := time.Now()

	if filter.StartDate != nil {
		startDate = *filter.StartDate
	}
	if filter.EndDate != nil {
		endDate = *filter.EndDate
	}

	metrics, total, err := s.systemMetricsRepo.GetByTenantAndDateRange(ctx, tenantID, startDate, endDate, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	responses := make([]*dtos.SystemUsageMetricsResponse, len(metrics))
	for i, metric := range metrics {
		responses[i] = s.toSystemMetricsResponse(metric)
	}

	return &dtos.SystemMetricsListResponse{
		Metrics:    responses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit))),
	}, nil
}

// GetSystemOverview retrieves system overview
func (s *ReportingAnalyticsServiceImpl) GetSystemOverview(ctx context.Context, tenantID string, days int) (*dtos.SystemOverviewResponse, error) {
	overview, err := s.systemMetricsRepo.GetSystemOverview(ctx, tenantID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get system overview: %w", err)
	}

	startDate := time.Now().AddDate(0, 0, -days)
	endDate := time.Now()

	response := &dtos.SystemOverviewResponse{
		TenantID:   tenantID,
		Period:     fmt.Sprintf("Last %d days", days),
		StartDate:  startDate,
		EndDate:    endDate,
		TopMetrics: []map[string]interface{}{},
		Trends:     []map[string]interface{}{},
		Alerts:     []map[string]interface{}{},
	}

	// Extract values from overview map
	if totalAPICalls, ok := overview["total_api_calls"].(int64); ok {
		response.TotalAPIsCalls = totalAPICalls
	}
	if successRate, ok := overview["success_rate"].(float64); ok {
		response.SuccessRate = successRate
	}
	if avgResponseTime, ok := overview["avg_response_time"].(float64); ok {
		response.AvgResponseTime = avgResponseTime
	}
	if storageUsed, ok := overview["storage_used"].(int64); ok {
		response.StorageUsed = storageUsed
	}
	if bandwidthUsed, ok := overview["bandwidth_used"].(int64); ok {
		response.BandwidthUsed = bandwidthUsed
	}
	if activeUsers, ok := overview["active_users"].(int); ok {
		response.ActiveUsers = activeUsers
	}
	if totalUsers, ok := overview["total_users"].(int); ok {
		response.TotalUsers = totalUsers
	}
	if newUsers, ok := overview["new_users"].(int); ok {
		response.NewUsers = newUsers
	}
	if totalSessions, ok := overview["total_sessions"].(int); ok {
		response.TotalSessions = totalSessions
	}
	if revenue, ok := overview["revenue"].(float64); ok {
		response.Revenue = revenue
	}
	if ordersCount, ok := overview["orders_count"].(int); ok {
		response.OrdersCount = ordersCount
	}

	return response, nil
}

// GetSystemStats retrieves specific system statistics
func (s *ReportingAnalyticsServiceImpl) GetSystemStats(ctx context.Context, tenantID string, metricType string, startDate, endDate time.Time) (map[string]interface{}, error) {
	switch metricType {
	case "api":
		return s.systemMetricsRepo.GetAPIUsageStats(ctx, tenantID, startDate, endDate)
	case "storage":
		return s.systemMetricsRepo.GetStorageStats(ctx, tenantID, startDate, endDate)
	case "users":
		return s.systemMetricsRepo.GetUserActivityStats(ctx, tenantID, startDate, endDate)
	case "performance":
		return s.systemMetricsRepo.GetPerformanceStats(ctx, tenantID, startDate, endDate)
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", metricType)
	}
}

// ============================
// Report Exports (Placeholder implementations)
// ============================

// CreateExport creates a new export request
func (s *ReportingAnalyticsServiceImpl) CreateExport(ctx context.Context, tenantID string, userID uuid.UUID, req *dtos.CreateReportExportRequest) (*dtos.ReportExportResponse, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("export functionality not yet implemented")
}

// GetExport retrieves an export by ID
func (s *ReportingAnalyticsServiceImpl) GetExport(ctx context.Context, tenantID string, exportID uuid.UUID) (*dtos.ReportExportResponse, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("export functionality not yet implemented")
}

// ListExports lists exports
func (s *ReportingAnalyticsServiceImpl) ListExports(ctx context.Context, tenantID string, userID *uuid.UUID, page, limit int) (*dtos.ExportListResponse, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("export functionality not yet implemented")
}

// DownloadExport downloads an export
func (s *ReportingAnalyticsServiceImpl) DownloadExport(ctx context.Context, tenantID string, exportID uuid.UUID) (string, error) {
	// Placeholder implementation
	return "", fmt.Errorf("export functionality not yet implemented")
}

// ============================
// Report Schedules (Placeholder implementations)
// ============================

// CreateSchedule creates a new scheduled report
func (s *ReportingAnalyticsServiceImpl) CreateSchedule(ctx context.Context, tenantID string, userID uuid.UUID, req *dtos.CreateReportScheduleRequest) (*dtos.ReportScheduleResponse, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("schedule functionality not yet implemented")
}

// GetSchedule retrieves a schedule by ID
func (s *ReportingAnalyticsServiceImpl) GetSchedule(ctx context.Context, tenantID string, scheduleID uuid.UUID) (*dtos.ReportScheduleResponse, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("schedule functionality not yet implemented")
}

// UpdateSchedule updates a schedule
func (s *ReportingAnalyticsServiceImpl) UpdateSchedule(ctx context.Context, tenantID string, scheduleID uuid.UUID, req *dtos.UpdateReportScheduleRequest) (*dtos.ReportScheduleResponse, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("schedule functionality not yet implemented")
}

// DeleteSchedule deletes a schedule
func (s *ReportingAnalyticsServiceImpl) DeleteSchedule(ctx context.Context, tenantID string, scheduleID uuid.UUID) error {
	// Placeholder implementation
	return fmt.Errorf("schedule functionality not yet implemented")
}

// ListSchedules lists schedules
func (s *ReportingAnalyticsServiceImpl) ListSchedules(ctx context.Context, tenantID string, userID *uuid.UUID, page, limit int) (*dtos.ScheduleListResponse, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("schedule functionality not yet implemented")
}

// ProcessScheduledReports processes due scheduled reports
func (s *ReportingAnalyticsServiceImpl) ProcessScheduledReports(ctx context.Context) error {
	// Placeholder implementation
	return fmt.Errorf("schedule processing not yet implemented")
}

// ============================
// Dashboard & Sales Reports (Placeholder implementations)
// ============================

// GetDashboardStats retrieves dashboard statistics
func (s *ReportingAnalyticsServiceImpl) GetDashboardStats(ctx context.Context, tenantID string, period string) (*dtos.DashboardStatsResponse, error) {
	// Placeholder implementation
	return &dtos.DashboardStatsResponse{
		TenantID:      tenantID,
		Period:        period,
		UserStats:     make(map[string]interface{}),
		SystemStats:   make(map[string]interface{}),
		SalesStats:    make(map[string]interface{}),
		RecentReports: []*dtos.AnalyticsReportResponse{},
		RecentExports: []*dtos.ReportExportResponse{},
		AlertsCount:   0,
		QuickMetrics:  []map[string]interface{}{},
		TrendData:     []map[string]interface{}{},
		GeneratedAt:   time.Now(),
	}, nil
}

// GenerateSalesReport generates a sales report
func (s *ReportingAnalyticsServiceImpl) GenerateSalesReport(ctx context.Context, tenantID string, userID uuid.UUID, reportType string, startDate, endDate time.Time) (*dtos.AnalyticsReportResponse, error) {
	// Placeholder implementation - would integrate with POS service
	return nil, fmt.Errorf("sales report generation not yet implemented")
}

// GetSalesStats retrieves sales statistics
func (s *ReportingAnalyticsServiceImpl) GetSalesStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Placeholder implementation - would integrate with POS service
	return make(map[string]interface{}), nil
}

// ============================
// Cleanup Operations
// ============================

// CleanupExpiredReports removes expired reports
func (s *ReportingAnalyticsServiceImpl) CleanupExpiredReports(ctx context.Context) error {
	s.logger.Info("Starting cleanup of expired reports")

	// Delete expired reports
	now := time.Now()
	deletedCount, err := s.reportRepo.DeleteExpiredReports(ctx, now)
	if err != nil {
		s.logger.Error("Failed to delete expired reports", zap.Error(err))
		return fmt.Errorf("failed to delete expired reports: %w", err)
	}

	// Delete expired exports
	exportDeletedCount, err := s.exportRepo.DeleteExpiredExports(ctx, now)
	if err != nil {
		s.logger.Error("Failed to delete expired exports", zap.Error(err))
		return fmt.Errorf("failed to delete expired exports: %w", err)
	}

	s.logger.Info("Cleanup completed",
		zap.Int64("reports_deleted", deletedCount),
		zap.Int64("exports_deleted", exportDeletedCount),
	)

	return nil
}

// CleanupOldMetrics removes old metrics based on retention policy
func (s *ReportingAnalyticsServiceImpl) CleanupOldMetrics(ctx context.Context, retentionDays int) error {
	s.logger.Info("Starting cleanup of old metrics", zap.Int("retention_days", retentionDays))

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Delete old activity metrics
	activityDeleted, err := s.activityRepo.DeleteOldMetrics(ctx, cutoffDate)
	if err != nil {
		s.logger.Error("Failed to delete old activity metrics", zap.Error(err))
		return fmt.Errorf("failed to delete old activity metrics: %w", err)
	}

	// Delete old system metrics
	systemDeleted, err := s.systemMetricsRepo.DeleteOldMetrics(ctx, cutoffDate)
	if err != nil {
		s.logger.Error("Failed to delete old system metrics", zap.Error(err))
		return fmt.Errorf("failed to delete old system metrics: %w", err)
	}

	s.logger.Info("Metrics cleanup completed",
		zap.Int64("activity_metrics_deleted", activityDeleted),
		zap.Int64("system_metrics_deleted", systemDeleted),
	)

	return nil
}

// ============================
// Helper Methods
// ============================

// toActivityMetricsResponse converts domain model to DTO
func (s *ReportingAnalyticsServiceImpl) toActivityMetricsResponse(metric *domain.UserActivityMetrics) *dtos.UserActivityMetricsResponse {
	return &dtos.UserActivityMetricsResponse{
		ID:                 metric.ID,
		UserID:             metric.UserID,
		Date:               metric.Date,
		SessionID:          metric.SessionID,
		PageViews:          metric.PageViews,
		UniquePages:        metric.UniquePages,
		SessionDuration:    metric.SessionDuration,
		ActionsCount:       metric.ActionsCount,
		LoginCount:         metric.LoginCount,
		LastSeenAt:         metric.LastSeenAt,
		DeviceType:         metric.DeviceType,
		Browser:            metric.Browser,
		Platform:           metric.Platform,
		Country:            metric.Country,
		Region:             metric.Region,
		City:               metric.City,
		ActivityData:       metric.ActivityData,
		PerformanceMetrics: metric.PerformanceMetrics,
		ErrorsCount:        metric.ErrorsCount,
		CreatedAt:          metric.CreatedAt,
		UpdatedAt:          metric.UpdatedAt,
	}
}

// toSystemMetricsResponse converts domain model to DTO
func (s *ReportingAnalyticsServiceImpl) toSystemMetricsResponse(metric *domain.SystemUsageMetrics) *dtos.SystemUsageMetricsResponse {
	return &dtos.SystemUsageMetricsResponse{
		ID:                metric.ID,
		Date:              metric.Date,
		Hour:              metric.Hour,
		MetricType:        metric.MetricType,
		MetricName:        metric.MetricName,
		MetricValue:       metric.MetricValue,
		MetricUnit:        metric.MetricUnit,
		APICallsTotal:     metric.APICallsTotal,
		APICallsSuccess:   metric.APICallsSuccess,
		APICallsError:     metric.APICallsError,
		APIResponseTime:   metric.APIResponseTime,
		StorageUsed:       metric.StorageUsed,
		BandwidthUsed:     metric.BandwidthUsed,
		FilesUploaded:     metric.FilesUploaded,
		FilesDownloaded:   metric.FilesDownloaded,
		DatabaseQueries:   metric.DatabaseQueries,
		DatabaseQueryTime: metric.DatabaseQueryTime,
		ActiveUsers:       metric.ActiveUsers,
		NewUsers:          metric.NewUsers,
		SessionsCreated:   metric.SessionsCreated,
		LoginAttempts:     metric.LoginAttempts,
		LoginSuccessful:   metric.LoginSuccessful,
		OrdersCreated:     metric.OrdersCreated,
		Revenue:           metric.Revenue,
		PaymentsProcessed: metric.PaymentsProcessed,
		CustomMetrics:     metric.CustomMetrics,
		Metadata:          metric.Metadata,
		CreatedAt:         metric.CreatedAt,
		UpdatedAt:         metric.UpdatedAt,
	}
}
