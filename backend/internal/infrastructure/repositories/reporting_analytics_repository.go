package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// AnalyticsReportRepositoryImpl implements AnalyticsReportRepository
type AnalyticsReportRepositoryImpl struct {
	db *gorm.DB
}

// NewAnalyticsReportRepository creates a new analytics report repository
func NewAnalyticsReportRepository(db *gorm.DB) domain.AnalyticsReportRepository {
	return &AnalyticsReportRepositoryImpl{db: db}
}

func (r *AnalyticsReportRepositoryImpl) Create(ctx context.Context, report *domain.AnalyticsReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *AnalyticsReportRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.AnalyticsReport, error) {
	var report domain.AnalyticsReport
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&report).Error
	return &report, err
}

func (r *AnalyticsReportRepositoryImpl) Update(ctx context.Context, report *domain.AnalyticsReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

func (r *AnalyticsReportRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.AnalyticsReport{}, id).Error
}

func (r *AnalyticsReportRepositoryImpl) GetByTenantID(ctx context.Context, tenantID string, filter *domain.ReportFilter) ([]*domain.AnalyticsReport, int64, error) {
	var reports []*domain.AnalyticsReport
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("tenant_id = ?", tenantID)

	// Apply filters
	if filter.ReportType != nil && *filter.ReportType != "" {
		query = query.Where("report_type = ?", *filter.ReportType)
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (filter.Page - 1) * filter.Limit
	query = query.Order("created_at DESC").Offset(offset).Limit(filter.Limit)

	err := query.Find(&reports).Error
	return reports, total, err
}

func (r *AnalyticsReportRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status, errorMessage string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	return r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AnalyticsReportRepositoryImpl) MarkCompleted(ctx context.Context, id uuid.UUID, filePath, fileURL string, fileSize int64) error {
	updates := map[string]interface{}{
		"status":       "completed",
		"file_path":    filePath,
		"file_url":     fileURL,
		"file_size":    fileSize,
		"completed_at": time.Now(),
		"updated_at":   time.Now(),
	}
	return r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AnalyticsReportRepositoryImpl) MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) error {
	updates := map[string]interface{}{
		"status":        "failed",
		"error_message": errorMessage,
		"updated_at":    time.Now(),
	}
	return r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AnalyticsReportRepositoryImpl) IncrementDownloadCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("id = ?", id).Update("download_count", gorm.Expr("download_count + 1")).Error
}

func (r *AnalyticsReportRepositoryImpl) UpdateLastDownload(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("id = ?", id).Update("last_download_at", time.Now()).Error
}

func (r *AnalyticsReportRepositoryImpl) GetScheduledReports(ctx context.Context, before time.Time) ([]*domain.AnalyticsReport, error) {
	var reports []*domain.AnalyticsReport
	err := r.db.WithContext(ctx).Where("scheduled_for <= ? AND status = ?", before, "pending").Find(&reports).Error
	return reports, err
}

func (r *AnalyticsReportRepositoryImpl) GetRecurringReports(ctx context.Context, before time.Time) ([]*domain.AnalyticsReport, error) {
	var reports []*domain.AnalyticsReport
	err := r.db.WithContext(ctx).Where("is_recurring = true AND next_run_at <= ?", before).Find(&reports).Error
	return reports, err
}

func (r *AnalyticsReportRepositoryImpl) UpdateNextRun(ctx context.Context, id uuid.UUID, nextRun time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("id = ?", id).Update("next_run_at", nextRun).Error
}

func (r *AnalyticsReportRepositoryImpl) DeleteExpiredReports(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&domain.AnalyticsReport{})
	return result.RowsAffected, result.Error
}

func (r *AnalyticsReportRepositoryImpl) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.AnalyticsReport{}).Error
}

func (r *AnalyticsReportRepositoryImpl) UpdateProgress(ctx context.Context, id uuid.UUID, progress int, stats map[string]interface{}) error {
	updates := map[string]interface{}{
		"processing_stats": stats,
		"updated_at":       time.Now(),
	}
	return r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AnalyticsReportRepositoryImpl) UpdateFileInfo(ctx context.Context, id uuid.UUID, filePath, fileURL string, fileSize int64) error {
	updates := map[string]interface{}{
		"file_path":  filePath,
		"file_url":   fileURL,
		"file_size":  fileSize,
		"updated_at": time.Now(),
	}
	return r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AnalyticsReportRepositoryImpl) SearchReports(ctx context.Context, tenantID, query string, limit, offset int) ([]*domain.AnalyticsReport, int64, error) {
	var reports []*domain.AnalyticsReport
	var total int64

	searchQuery := r.db.WithContext(ctx).Model(&domain.AnalyticsReport{}).
		Where("tenant_id = ? AND (title ILIKE ? OR description ILIKE ?)", tenantID, "%"+query+"%", "%"+query+"%")

	// Count total
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := searchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Find(&reports).Error
	return reports, total, err
}

func (r *AnalyticsReportRepositoryImpl) GetReportsByType(ctx context.Context, tenantID, reportType string, limit, offset int) ([]*domain.AnalyticsReport, error) {
	var reports []*domain.AnalyticsReport
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND report_type = ?", tenantID, reportType).
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&reports).Error
	return reports, err
}

func (r *AnalyticsReportRepositoryImpl) GetRecentReports(ctx context.Context, tenantID string, limit int) ([]*domain.AnalyticsReport, error) {
	var reports []*domain.AnalyticsReport
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).
		Order("created_at DESC").Limit(limit).Find(&reports).Error
	return reports, err
}

// UserActivityMetricsRepositoryImpl implements UserActivityMetricsRepository
type UserActivityMetricsRepositoryImpl struct {
	db *gorm.DB
}

// NewUserActivityMetricsRepository creates a new user activity metrics repository
func NewUserActivityMetricsRepository(db *gorm.DB) domain.UserActivityMetricsRepository {
	return &UserActivityMetricsRepositoryImpl{db: db}
}

func (r *UserActivityMetricsRepositoryImpl) Create(ctx context.Context, metric *domain.UserActivityMetrics) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

func (r *UserActivityMetricsRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserActivityMetrics, error) {
	var metric domain.UserActivityMetrics
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&metric).Error
	return &metric, err
}

func (r *UserActivityMetricsRepositoryImpl) GetByTenantAndDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, filter *domain.ActivityMetricsFilter) ([]*domain.UserActivityMetrics, int64, error) {
	var metrics []*domain.UserActivityMetrics
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate)

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.DeviceType != nil && *filter.DeviceType != "" {
		query = query.Where("device_type = ?", *filter.DeviceType)
	}
	if filter.Country != nil && *filter.Country != "" {
		query = query.Where("country = ?", *filter.Country)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (filter.Page - 1) * filter.Limit
	query = query.Order("date DESC, created_at DESC").Offset(offset).Limit(filter.Limit)

	err := query.Find(&metrics).Error
	return metrics, total, err
}

func (r *UserActivityMetricsRepositoryImpl) RecordUserActivity(ctx context.Context, tenantID string, userID uuid.UUID, sessionID string, activityData map[string]interface{}) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Try to find existing metric for today
	var metric domain.UserActivityMetrics
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND date = ? AND session_id = ?",
		tenantID, userID, today, sessionID).First(&metric).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new metric
			metric = domain.UserActivityMetrics{
				TenantID:           tenantID,
				UserID:             userID,
				Date:               today,
				SessionID:          sessionID,
				PageViews:          1,
				UniquePages:        1,
				SessionDuration:    0,
				ActionsCount:       1,
				LoginCount:         0,
				LastSeenAt:         &[]time.Time{time.Now()}[0],
				ActivityData:       activityData,
				PerformanceMetrics: make(map[string]interface{}),
				ErrorsCount:        0,
			}

			// Extract additional data from activityData
			if deviceType, ok := activityData["device_type"].(string); ok {
				metric.DeviceType = deviceType
			}
			if browser, ok := activityData["browser"].(string); ok {
				metric.Browser = browser
			}
			if platform, ok := activityData["platform"].(string); ok {
				metric.Platform = platform
			}
			if country, ok := activityData["country"].(string); ok {
				metric.Country = country
			}
			if region, ok := activityData["region"].(string); ok {
				metric.Region = region
			}
			if city, ok := activityData["city"].(string); ok {
				metric.City = city
			}

			return r.db.WithContext(ctx).Create(&metric).Error
		}
		return err
	}

	// Update existing metric
	updates := map[string]interface{}{
		"page_views":    gorm.Expr("page_views + 1"),
		"actions_count": gorm.Expr("actions_count + 1"),
		"last_seen_at":  time.Now(),
		"updated_at":    time.Now(),
	}

	// Update activity data by merging
	if len(activityData) > 0 {
		// This is a simplified merge - in production, you might want more sophisticated merging
		if metric.ActivityData == nil {
			metric.ActivityData = make(map[string]interface{})
		}
		for k, v := range activityData {
			metric.ActivityData[k] = v
		}
		updates["activity_data"] = metric.ActivityData
	}

	return r.db.WithContext(ctx).Model(&metric).Where("id = ?", metric.ID).Updates(updates).Error
}

func (r *UserActivityMetricsRepositoryImpl) GetUserSummary(ctx context.Context, tenantID string, userID uuid.UUID, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	var result struct {
		TotalPageViews int     `json:"total_page_views"`
		TotalSessions  int     `json:"total_sessions"`
		TotalDuration  int     `json:"total_duration"`
		TotalActions   int     `json:"total_actions"`
		LoginCount     int     `json:"login_count"`
		AvgSessionTime float64 `json:"avg_session_time"`
	}

	err := r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Select("SUM(page_views) as total_page_views, COUNT(*) as total_sessions, SUM(session_duration) as total_duration, SUM(actions_count) as total_actions, SUM(login_count) as login_count, AVG(session_duration) as avg_session_time").
		Where("tenant_id = ? AND user_id = ? AND date >= ?", tenantID, userID, startDate).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_page_views": result.TotalPageViews,
		"total_sessions":   result.TotalSessions,
		"total_duration":   result.TotalDuration,
		"avg_session_time": int(result.AvgSessionTime),
		"total_actions":    result.TotalActions,
		"login_count":      result.LoginCount,
	}, nil
}

func (r *UserActivityMetricsRepositoryImpl) GetActivityTrends(ctx context.Context, tenantID string, startDate, endDate time.Time, groupBy string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	var groupFormat string
	switch groupBy {
	case "hour":
		groupFormat = "YYYY-MM-DD HH24:00:00"
	case "day":
		groupFormat = "YYYY-MM-DD"
	case "week":
		groupFormat = "YYYY-\"W\"IW"
	case "month":
		groupFormat = "YYYY-MM"
	default:
		groupFormat = "YYYY-MM-DD"
	}

	query := fmt.Sprintf(`
		SELECT 
			TO_CHAR(date, '%s') as period,
			SUM(page_views) as total_page_views,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(*) as total_sessions,
			SUM(actions_count) as total_actions
		FROM user_activity_metrics 
		WHERE tenant_id = ? AND date BETWEEN ? AND ?
		GROUP BY TO_CHAR(date, '%s')
		ORDER BY period
	`, groupFormat, groupFormat)

	rows, err := r.db.WithContext(ctx).Raw(query, tenantID, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var period string
		var totalPageViews, uniqueUsers, totalSessions, totalActions int

		if err := rows.Scan(&period, &totalPageViews, &uniqueUsers, &totalSessions, &totalActions); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"period":           period,
			"total_page_views": totalPageViews,
			"unique_users":     uniqueUsers,
			"total_sessions":   totalSessions,
			"total_actions":    totalActions,
		})
	}

	return results, nil
}

func (r *UserActivityMetricsRepositoryImpl) DeleteOldMetrics(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("date < ?", before).Delete(&domain.UserActivityMetrics{})
	return result.RowsAffected, result.Error
}

// SystemUsageMetricsRepositoryImpl implements SystemUsageMetricsRepository
type SystemUsageMetricsRepositoryImpl struct {
	db *gorm.DB
}

// NewSystemUsageMetricsRepository creates a new system usage metrics repository
func NewSystemUsageMetricsRepository(db *gorm.DB) domain.SystemUsageMetricsRepository {
	return &SystemUsageMetricsRepositoryImpl{db: db}
}

func (r *SystemUsageMetricsRepositoryImpl) Create(ctx context.Context, metric *domain.SystemUsageMetrics) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

func (r *SystemUsageMetricsRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.SystemUsageMetrics, error) {
	var metric domain.SystemUsageMetrics
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&metric).Error
	return &metric, err
}

func (r *SystemUsageMetricsRepositoryImpl) GetByTenantAndDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, filter *domain.SystemMetricsFilter) ([]*domain.SystemUsageMetrics, int64, error) {
	var metrics []*domain.SystemUsageMetrics
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate)

	// Apply filters
	if filter.MetricType != nil && *filter.MetricType != "" {
		query = query.Where("metric_type = ?", *filter.MetricType)
	}
	if filter.MetricName != nil && *filter.MetricName != "" {
		query = query.Where("metric_name = ?", *filter.MetricName)
	}
	if filter.Hour != nil {
		query = query.Where("hour = ?", *filter.Hour)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (filter.Page - 1) * filter.Limit
	query = query.Order("date DESC, hour DESC").Offset(offset).Limit(filter.Limit)

	err := query.Find(&metrics).Error
	return metrics, total, err
}

func (r *SystemUsageMetricsRepositoryImpl) RecordCustomMetric(ctx context.Context, tenantID string, date time.Time, hour int, metricType, metricName string, metricValue float64, metricUnit string, customMetrics map[string]interface{}) error {
	// Try to find existing metric for this date/hour/type/name
	var metric domain.SystemUsageMetrics
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND date = ? AND hour = ? AND metric_type = ? AND metric_name = ?",
		tenantID, date, hour, metricType, metricName).First(&metric).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new metric
			metric = domain.SystemUsageMetrics{
				TenantID:      tenantID,
				Date:          date,
				Hour:          hour,
				MetricType:    metricType,
				MetricName:    metricName,
				MetricValue:   metricValue,
				MetricUnit:    metricUnit,
				CustomMetrics: customMetrics,
				Metadata:      make(map[string]interface{}),
			}
			return r.db.WithContext(ctx).Create(&metric).Error
		}
		return err
	}

	// Update existing metric (accumulate values)
	updates := map[string]interface{}{
		"metric_value": gorm.Expr("metric_value + ?", metricValue),
		"updated_at":   time.Now(),
	}

	// Update custom metrics by merging
	if len(customMetrics) > 0 {
		if metric.CustomMetrics == nil {
			metric.CustomMetrics = make(map[string]interface{})
		}
		for k, v := range customMetrics {
			metric.CustomMetrics[k] = v
		}
		updates["custom_metrics"] = metric.CustomMetrics
	}

	return r.db.WithContext(ctx).Model(&metric).Where("id = ?", metric.ID).Updates(updates).Error
}

func (r *SystemUsageMetricsRepositoryImpl) GetSystemOverview(ctx context.Context, tenantID string, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	// Get aggregated metrics
	var apiMetrics struct {
		TotalAPICalls int64   `json:"total_api_calls"`
		TotalErrors   int64   `json:"total_errors"`
		AvgResponse   float64 `json:"avg_response_time"`
	}

	err := r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("SUM(api_calls_total) as total_api_calls, SUM(api_calls_error) as total_errors, AVG(api_response_time) as avg_response").
		Where("tenant_id = ? AND date >= ?", tenantID, startDate).
		Scan(&apiMetrics).Error

	if err != nil {
		return nil, err
	}

	// Calculate success rate
	successRate := float64(100)
	if apiMetrics.TotalAPICalls > 0 {
		successRate = float64(apiMetrics.TotalAPICalls-apiMetrics.TotalErrors) / float64(apiMetrics.TotalAPICalls) * 100
	}

	// Get storage and bandwidth totals
	var resourceMetrics struct {
		StorageUsed   int64 `json:"storage_used"`
		BandwidthUsed int64 `json:"bandwidth_used"`
	}

	err = r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("SUM(storage_used) as storage_used, SUM(bandwidth_used) as bandwidth_used").
		Where("tenant_id = ? AND date >= ?", tenantID, startDate).
		Scan(&resourceMetrics).Error

	if err != nil {
		return nil, err
	}

	// Get user metrics
	var userMetrics struct {
		ActiveUsers   int `json:"active_users"`
		NewUsers      int `json:"new_users"`
		TotalUsers    int `json:"total_users"`
		TotalSessions int `json:"total_sessions"`
	}

	err = r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("MAX(active_users) as active_users, SUM(new_users) as new_users, MAX(active_users) as total_users, SUM(sessions_created) as total_sessions").
		Where("tenant_id = ? AND date >= ?", tenantID, startDate).
		Scan(&userMetrics).Error

	if err != nil {
		return nil, err
	}

	// Get sales metrics
	var salesMetrics struct {
		Revenue     float64 `json:"revenue"`
		OrdersCount int     `json:"orders_count"`
	}

	err = r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("SUM(revenue) as revenue, SUM(orders_created) as orders_count").
		Where("tenant_id = ? AND date >= ?", tenantID, startDate).
		Scan(&salesMetrics).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_api_calls":   apiMetrics.TotalAPICalls,
		"success_rate":      successRate,
		"avg_response_time": apiMetrics.AvgResponse,
		"storage_used":      resourceMetrics.StorageUsed,
		"bandwidth_used":    resourceMetrics.BandwidthUsed,
		"active_users":      userMetrics.ActiveUsers,
		"total_users":       userMetrics.TotalUsers,
		"new_users":         userMetrics.NewUsers,
		"total_sessions":    userMetrics.TotalSessions,
		"revenue":           salesMetrics.Revenue,
		"orders_count":      salesMetrics.OrdersCount,
	}, nil
}

func (r *SystemUsageMetricsRepositoryImpl) GetAPIUsageStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalCalls   int64   `json:"total_calls"`
		SuccessCalls int64   `json:"success_calls"`
		ErrorCalls   int64   `json:"error_calls"`
		AvgResponse  float64 `json:"avg_response_time"`
		MaxResponse  int     `json:"max_response_time"`
		MinResponse  int     `json:"min_response_time"`
	}

	err := r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("SUM(api_calls_total) as total_calls, SUM(api_calls_success) as success_calls, SUM(api_calls_error) as error_calls, AVG(api_response_time) as avg_response, MAX(api_response_time) as max_response, MIN(api_response_time) as min_response").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	errorRate := float64(0)
	if stats.TotalCalls > 0 {
		errorRate = float64(stats.ErrorCalls) / float64(stats.TotalCalls) * 100
	}

	return map[string]interface{}{
		"total_calls":       stats.TotalCalls,
		"success_calls":     stats.SuccessCalls,
		"error_calls":       stats.ErrorCalls,
		"error_rate":        errorRate,
		"avg_response_time": stats.AvgResponse,
		"max_response_time": stats.MaxResponse,
		"min_response_time": stats.MinResponse,
	}, nil
}

func (r *SystemUsageMetricsRepositoryImpl) GetStorageStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalStorage    int64 `json:"total_storage"`
		FilesUploaded   int64 `json:"files_uploaded"`
		FilesDownloaded int64 `json:"files_downloaded"`
		BandwidthUsed   int64 `json:"bandwidth_used"`
	}

	err := r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("MAX(storage_used) as total_storage, SUM(files_uploaded) as files_uploaded, SUM(files_downloaded) as files_downloaded, SUM(bandwidth_used) as bandwidth_used").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_storage":    stats.TotalStorage,
		"files_uploaded":   stats.FilesUploaded,
		"files_downloaded": stats.FilesDownloaded,
		"bandwidth_used":   stats.BandwidthUsed,
	}, nil
}

func (r *SystemUsageMetricsRepositoryImpl) GetUserActivityStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var stats struct {
		MaxActiveUsers   int `json:"max_active_users"`
		TotalNewUsers    int `json:"total_new_users"`
		TotalSessions    int `json:"total_sessions"`
		TotalLogins      int `json:"total_logins"`
		SuccessfulLogins int `json:"successful_logins"`
	}

	err := r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("MAX(active_users) as max_active_users, SUM(new_users) as total_new_users, SUM(sessions_created) as total_sessions, SUM(login_attempts) as total_logins, SUM(login_successful) as successful_logins").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	loginSuccessRate := float64(0)
	if stats.TotalLogins > 0 {
		loginSuccessRate = float64(stats.SuccessfulLogins) / float64(stats.TotalLogins) * 100
	}

	return map[string]interface{}{
		"max_active_users":   stats.MaxActiveUsers,
		"total_new_users":    stats.TotalNewUsers,
		"total_sessions":     stats.TotalSessions,
		"total_logins":       stats.TotalLogins,
		"successful_logins":  stats.SuccessfulLogins,
		"login_success_rate": loginSuccessRate,
	}, nil
}

func (r *SystemUsageMetricsRepositoryImpl) GetPerformanceStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var stats struct {
		AvgAPIResponse float64 `json:"avg_api_response"`
		AvgDBQuery     float64 `json:"avg_db_query"`
		TotalDBQueries int64   `json:"total_db_queries"`
	}

	err := r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("AVG(api_response_time) as avg_api_response, AVG(database_query_time) as avg_db_query, SUM(database_queries) as total_db_queries").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"avg_api_response_time": stats.AvgAPIResponse,
		"avg_db_query_time":     stats.AvgDBQuery,
		"total_db_queries":      stats.TotalDBQueries,
	}, nil
}

func (r *SystemUsageMetricsRepositoryImpl) DeleteOldMetrics(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("date < ?", before).Delete(&domain.SystemUsageMetrics{})
	return result.RowsAffected, result.Error
}
