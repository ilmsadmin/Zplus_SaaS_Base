package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// =============================================
// MISSING METHODS FOR USER ACTIVITY METRICS
// =============================================

func (r *UserActivityMetricsRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserActivityMetrics{}, id).Error
}

func (r *UserActivityMetricsRepositoryImpl) Update(ctx context.Context, metrics *domain.UserActivityMetrics) error {
	return r.db.WithContext(ctx).Save(metrics).Error
}

func (r *UserActivityMetricsRepositoryImpl) UpdateSessionDuration(ctx context.Context, tenantID string, userID uuid.UUID, sessionID string, duration int) error {
	return r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Where("tenant_id = ? AND user_id = ? AND session_id = ?", tenantID, userID, sessionID).
		Update("session_duration", duration).Error
}

func (r *UserActivityMetricsRepositoryImpl) IncrementPageViews(ctx context.Context, tenantID string, userID uuid.UUID, date time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Where("tenant_id = ? AND user_id = ? AND date = ?", tenantID, userID, date).
		Update("page_views", gorm.Expr("page_views + 1")).Error
}

func (r *UserActivityMetricsRepositoryImpl) IncrementActions(ctx context.Context, tenantID string, userID uuid.UUID, date time.Time, count int) error {
	return r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Where("tenant_id = ? AND user_id = ? AND date = ?", tenantID, userID, date).
		Update("actions_count", gorm.Expr("actions_count + ?", count)).Error
}

func (r *UserActivityMetricsRepositoryImpl) IncrementLoginCount(ctx context.Context, tenantID string, userID uuid.UUID, date time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Where("tenant_id = ? AND user_id = ? AND date = ?", tenantID, userID, date).
		Update("login_count", gorm.Expr("login_count + 1")).Error
}

func (r *UserActivityMetricsRepositoryImpl) IncrementErrorCount(ctx context.Context, tenantID string, userID uuid.UUID, date time.Time, count int) error {
	return r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Where("tenant_id = ? AND user_id = ? AND date = ?", tenantID, userID, date).
		Update("errors_count", gorm.Expr("errors_count + ?", count)).Error
}

func (r *UserActivityMetricsRepositoryImpl) GetByTenantAndUser(ctx context.Context, tenantID string, userID uuid.UUID, filter *domain.ActivityMetricsFilter) ([]*domain.UserActivityMetrics, int64, error) {
	var metrics []*domain.UserActivityMetrics
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID)

	// Apply date filters if provided
	if filter.StartDate != nil {
		query = query.Where("date >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("date <= ?", *filter.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (filter.Page - 1) * filter.Limit
	query = query.Order("date DESC").Offset(offset).Limit(filter.Limit)

	err := query.Find(&metrics).Error
	return metrics, total, err
}

func (r *UserActivityMetricsRepositoryImpl) GetDailyMetrics(ctx context.Context, tenantID string, date time.Time, filter *domain.ActivityMetricsFilter) ([]*domain.UserActivityMetrics, error) {
	var metrics []*domain.UserActivityMetrics
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND date = ?", tenantID, date)

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	err := query.Find(&metrics).Error
	return metrics, err
}

func (r *UserActivityMetricsRepositoryImpl) GetActiveUsersCount(ctx context.Context, tenantID string, date time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Where("tenant_id = ? AND date = ?", tenantID, date).
		Select("COUNT(DISTINCT user_id)").
		Count(&count).Error
	return count, err
}

func (r *UserActivityMetricsRepositoryImpl) GetTopUsers(ctx context.Context, tenantID string, startDate, endDate time.Time, limit int) ([]*domain.UserActivityMetrics, error) {
	var metrics []*domain.UserActivityMetrics
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Order("page_views DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

func (r *UserActivityMetricsRepositoryImpl) GetDeviceStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	type DeviceStat struct {
		DeviceType string `json:"device_type"`
		Count      int64  `json:"count"`
	}

	var stats []DeviceStat
	err := r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Select("device_type, COUNT(*) as count").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Group("device_type").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, stat := range stats {
		result[stat.DeviceType] = stat.Count
	}
	return result, nil
}

func (r *UserActivityMetricsRepositoryImpl) GetGeographicStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	type GeoStat struct {
		Country string `json:"country"`
		Count   int64  `json:"count"`
	}

	var stats []GeoStat
	err := r.db.WithContext(ctx).Model(&domain.UserActivityMetrics{}).
		Select("country, COUNT(*) as count").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Group("country").
		Order("count DESC").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, stat := range stats {
		result[stat.Country] = stat.Count
	}
	return result, nil
}

func (r *UserActivityMetricsRepositoryImpl) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.UserActivityMetrics{}).Error
}

// =============================================
// MISSING METHODS FOR SYSTEM USAGE METRICS
// =============================================

func (r *SystemUsageMetricsRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.SystemUsageMetrics{}, id).Error
}

func (r *SystemUsageMetricsRepositoryImpl) Update(ctx context.Context, metrics *domain.SystemUsageMetrics) error {
	return r.db.WithContext(ctx).Save(metrics).Error
}

func (r *SystemUsageMetricsRepositoryImpl) RecordAPIUsage(ctx context.Context, tenantID string, date time.Time, hour int, calls, successes, errors int, avgResponseTime int) error {
	return r.RecordCustomMetric(ctx, tenantID, date, hour, "api", "usage", float64(calls), "calls", map[string]interface{}{
		"successes":         successes,
		"errors":            errors,
		"avg_response_time": avgResponseTime,
	})
}

func (r *SystemUsageMetricsRepositoryImpl) RecordStorageUsage(ctx context.Context, tenantID string, date time.Time, hour int, storageUsed, bandwidthUsed int64, filesUp, filesDown int) error {
	return r.RecordCustomMetric(ctx, tenantID, date, hour, "storage", "usage", float64(storageUsed), "bytes", map[string]interface{}{
		"bandwidth_used":   bandwidthUsed,
		"files_uploaded":   filesUp,
		"files_downloaded": filesDown,
	})
}

func (r *SystemUsageMetricsRepositoryImpl) RecordDatabaseUsage(ctx context.Context, tenantID string, date time.Time, hour int, queries int, totalQueryTime int) error {
	return r.RecordCustomMetric(ctx, tenantID, date, hour, "database", "usage", float64(queries), "queries", map[string]interface{}{
		"total_query_time": totalQueryTime,
		"avg_query_time":   float64(totalQueryTime) / float64(queries),
	})
}

func (r *SystemUsageMetricsRepositoryImpl) RecordUserMetrics(ctx context.Context, tenantID string, date time.Time, hour int, activeUsers, newUsers, sessions, logins, loginSuccesses int) error {
	return r.RecordCustomMetric(ctx, tenantID, date, hour, "user", "activity", float64(activeUsers), "users", map[string]interface{}{
		"new_users":       newUsers,
		"sessions":        sessions,
		"logins":          logins,
		"login_successes": loginSuccesses,
	})
}

func (r *SystemUsageMetricsRepositoryImpl) RecordPOSMetrics(ctx context.Context, tenantID string, date time.Time, hour int, orders int, revenue float64, payments int) error {
	return r.RecordCustomMetric(ctx, tenantID, date, hour, "pos", "sales", revenue, "currency", map[string]interface{}{
		"orders":   orders,
		"payments": payments,
	})
}

func (r *SystemUsageMetricsRepositoryImpl) GetHourlyMetrics(ctx context.Context, tenantID string, date time.Time, filter *domain.SystemMetricsFilter) ([]*domain.SystemUsageMetrics, error) {
	var metrics []*domain.SystemUsageMetrics
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND date = ?", tenantID, date)

	if filter.MetricType != nil && *filter.MetricType != "" {
		query = query.Where("metric_type = ?", *filter.MetricType)
	}
	if filter.Hour != nil {
		query = query.Where("hour = ?", *filter.Hour)
	}

	err := query.Order("hour ASC").Find(&metrics).Error
	return metrics, err
}

func (r *SystemUsageMetricsRepositoryImpl) GetDailyAggregates(ctx context.Context, tenantID string, startDate, endDate time.Time, metricTypes []string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT 
			date,
			metric_type,
			SUM(metric_value) as total_value,
			AVG(metric_value) as avg_value,
			MAX(metric_value) as max_value,
			MIN(metric_value) as min_value,
			COUNT(*) as data_points
		FROM system_usage_metrics 
		WHERE tenant_id = ? AND date BETWEEN ? AND ?
	`

	args := []interface{}{tenantID, startDate, endDate}

	if len(metricTypes) > 0 {
		query += " AND metric_type IN (?)"
		args = append(args, metricTypes)
	}

	query += " GROUP BY date, metric_type ORDER BY date, metric_type"

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var date time.Time
		var metricType string
		var totalValue, avgValue, maxValue, minValue float64
		var dataPoints int

		if err := rows.Scan(&date, &metricType, &totalValue, &avgValue, &maxValue, &minValue, &dataPoints); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"date":        date,
			"metric_type": metricType,
			"total_value": totalValue,
			"avg_value":   avgValue,
			"max_value":   maxValue,
			"min_value":   minValue,
			"data_points": dataPoints,
		})
	}

	return results, nil
}

func (r *SystemUsageMetricsRepositoryImpl) GetTopMetrics(ctx context.Context, tenantID string, metricType string, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT 
			metric_name,
			SUM(metric_value) as total_value,
			AVG(metric_value) as avg_value,
			COUNT(*) as data_points
		FROM system_usage_metrics 
		WHERE tenant_id = ? AND metric_type = ? AND date BETWEEN ? AND ?
		GROUP BY metric_name 
		ORDER BY total_value DESC 
		LIMIT ?
	`

	rows, err := r.db.WithContext(ctx).Raw(query, tenantID, metricType, startDate, endDate, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var metricName string
		var totalValue, avgValue float64
		var dataPoints int

		if err := rows.Scan(&metricName, &totalValue, &avgValue, &dataPoints); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"metric_name": metricName,
			"total_value": totalValue,
			"avg_value":   avgValue,
			"data_points": dataPoints,
		})
	}

	return results, nil
}

func (r *SystemUsageMetricsRepositoryImpl) GetSystemWideStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalTenants     int64   `json:"total_tenants"`
		TotalAPICalls    int64   `json:"total_api_calls"`
		TotalStorageUsed int64   `json:"total_storage_used"`
		TotalRevenue     float64 `json:"total_revenue"`
		ActiveTenants    int64   `json:"active_tenants"`
	}

	err := r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("COUNT(DISTINCT tenant_id) as total_tenants, SUM(api_calls_total) as total_api_calls, SUM(storage_used) as total_storage_used, SUM(revenue) as total_revenue").
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// Get active tenants (those with activity in the period)
	err = r.db.WithContext(ctx).Model(&domain.SystemUsageMetrics{}).
		Select("COUNT(DISTINCT tenant_id)").
		Where("date BETWEEN ? AND ? AND (api_calls_total > 0 OR orders_created > 0)", startDate, endDate).
		Scan(&stats.ActiveTenants).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_tenants":      stats.TotalTenants,
		"active_tenants":     stats.ActiveTenants,
		"total_api_calls":    stats.TotalAPICalls,
		"total_storage_used": stats.TotalStorageUsed,
		"total_revenue":      stats.TotalRevenue,
	}, nil
}

func (r *SystemUsageMetricsRepositoryImpl) GetTenantRankings(ctx context.Context, metricName string, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// Map metric names to actual database columns
	var column string
	switch metricName {
	case "api_calls":
		column = "api_calls_total"
	case "storage":
		column = "storage_used"
	case "revenue":
		column = "revenue"
	case "users":
		column = "active_users"
	default:
		column = "metric_value"
	}

	query := fmt.Sprintf(`
		SELECT 
			tenant_id,
			SUM(%s) as total_value,
			AVG(%s) as avg_value,
			MAX(%s) as max_value
		FROM system_usage_metrics 
		WHERE date BETWEEN ? AND ?
		GROUP BY tenant_id 
		ORDER BY total_value DESC 
		LIMIT ?
	`, column, column, column)

	rows, err := r.db.WithContext(ctx).Raw(query, startDate, endDate, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rank := 1
	for rows.Next() {
		var tenantID string
		var totalValue, avgValue, maxValue float64

		if err := rows.Scan(&tenantID, &totalValue, &avgValue, &maxValue); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"rank":        rank,
			"tenant_id":   tenantID,
			"total_value": totalValue,
			"avg_value":   avgValue,
			"max_value":   maxValue,
		})
		rank++
	}

	return results, nil
}

func (r *SystemUsageMetricsRepositoryImpl) DeleteByTenantID(ctx context.Context, tenantID string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&domain.SystemUsageMetrics{}).Error
}
