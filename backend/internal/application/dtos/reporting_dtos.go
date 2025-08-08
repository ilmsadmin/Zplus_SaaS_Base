package dtos

import (
	"time"

	"github.com/google/uuid"
)

// ============================
// Analytics Report DTOs
// ============================

// CreateAnalyticsReportRequest for creating new reports
type CreateAnalyticsReportRequest struct {
	ReportType     string                 `json:"report_type" validate:"required,oneof=sales users system inventory financial"`
	ReportSubtype  string                 `json:"report_subtype" validate:"required,oneof=daily weekly monthly yearly custom"`
	Title          string                 `json:"title" validate:"required,min=3,max=255"`
	Description    string                 `json:"description,omitempty"`
	PeriodStart    time.Time              `json:"period_start" validate:"required"`
	PeriodEnd      time.Time              `json:"period_end" validate:"required,gtfield=PeriodStart"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	FileFormat     string                 `json:"file_format" validate:"oneof=json pdf excel csv"`
	ScheduledFor   *time.Time             `json:"scheduled_for,omitempty"`
	IsRecurring    bool                   `json:"is_recurring"`
	RecurrenceRule string                 `json:"recurrence_rule,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateAnalyticsReportRequest for updating reports
type UpdateAnalyticsReportRequest struct {
	Title          *string                `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
	Description    *string                `json:"description,omitempty"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	FileFormat     *string                `json:"file_format,omitempty" validate:"omitempty,oneof=json pdf excel csv"`
	IsRecurring    *bool                  `json:"is_recurring,omitempty"`
	RecurrenceRule *string                `json:"recurrence_rule,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsReportResponse for report responses
type AnalyticsReportResponse struct {
	ID              uuid.UUID              `json:"id"`
	ReportType      string                 `json:"report_type"`
	ReportSubtype   string                 `json:"report_subtype"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	Parameters      map[string]interface{} `json:"parameters"`
	Summary         map[string]interface{} `json:"summary"`
	FileURL         string                 `json:"file_url,omitempty"`
	FileFormat      string                 `json:"file_format"`
	FileSize        int64                  `json:"file_size"`
	Status          string                 `json:"status"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	ScheduledFor    *time.Time             `json:"scheduled_for,omitempty"`
	ProcessingStats map[string]interface{} `json:"processing_stats"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	DownloadCount   int                    `json:"download_count"`
	LastDownloadAt  *time.Time             `json:"last_download_at,omitempty"`
	IsRecurring     bool                   `json:"is_recurring"`
	RecurrenceRule  string                 `json:"recurrence_rule,omitempty"`
	NextRunAt       *time.Time             `json:"next_run_at,omitempty"`
	Tags            []string               `json:"tags"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
}

// ReportListResponse for paginated report lists
type ReportListResponse struct {
	Reports    []*AnalyticsReportResponse `json:"reports"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	Limit      int                        `json:"limit"`
	TotalPages int                        `json:"total_pages"`
}

// ============================
// User Activity DTOs
// ============================

// RecordUserActivityRequest for recording user activity
type RecordUserActivityRequest struct {
	UserID             uuid.UUID              `json:"user_id" validate:"required"`
	SessionID          string                 `json:"session_id" validate:"required"`
	Page               string                 `json:"page,omitempty"`
	Action             string                 `json:"action,omitempty"`
	DeviceType         string                 `json:"device_type,omitempty"`
	Browser            string                 `json:"browser,omitempty"`
	Platform           string                 `json:"platform,omitempty"`
	Country            string                 `json:"country,omitempty"`
	Region             string                 `json:"region,omitempty"`
	City               string                 `json:"city,omitempty"`
	IPAddress          string                 `json:"ip_address,omitempty"`
	ReferrerDomain     string                 `json:"referrer_domain,omitempty"`
	ReferrerURL        string                 `json:"referrer_url,omitempty"`
	ActivityData       map[string]interface{} `json:"activity_data,omitempty"`
	PerformanceMetrics map[string]interface{} `json:"performance_metrics,omitempty"`
}

// UserActivityMetricsResponse for activity metrics responses
type UserActivityMetricsResponse struct {
	ID                 uuid.UUID              `json:"id"`
	UserID             uuid.UUID              `json:"user_id"`
	Date               time.Time              `json:"date"`
	SessionID          string                 `json:"session_id"`
	PageViews          int                    `json:"page_views"`
	UniquePages        int                    `json:"unique_pages"`
	SessionDuration    int                    `json:"session_duration"`
	ActionsCount       int                    `json:"actions_count"`
	LoginCount         int                    `json:"login_count"`
	LastSeenAt         *time.Time             `json:"last_seen_at,omitempty"`
	DeviceType         string                 `json:"device_type"`
	Browser            string                 `json:"browser"`
	Platform           string                 `json:"platform"`
	Country            string                 `json:"country"`
	Region             string                 `json:"region"`
	City               string                 `json:"city"`
	ActivityData       map[string]interface{} `json:"activity_data"`
	PerformanceMetrics map[string]interface{} `json:"performance_metrics"`
	ErrorsCount        int                    `json:"errors_count"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// UserActivitySummaryResponse for user activity summaries
type UserActivitySummaryResponse struct {
	UserID          uuid.UUID                     `json:"user_id"`
	TotalPageViews  int                           `json:"total_page_views"`
	TotalSessions   int                           `json:"total_sessions"`
	TotalDuration   int                           `json:"total_duration"` // in seconds
	AvgSessionTime  int                           `json:"avg_session_time"`
	TotalActions    int                           `json:"total_actions"`
	LoginCount      int                           `json:"login_count"`
	LastActivity    *time.Time                    `json:"last_activity,omitempty"`
	TopPages        []map[string]interface{}      `json:"top_pages"`
	DeviceBreakdown map[string]int                `json:"device_breakdown"`
	DailyActivities []UserActivityMetricsResponse `json:"daily_activities"`
}

// ActivityMetricsListResponse for paginated activity metrics
type ActivityMetricsListResponse struct {
	Metrics    []*UserActivityMetricsResponse `json:"metrics"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	Limit      int                            `json:"limit"`
	TotalPages int                            `json:"total_pages"`
}

// ============================
// System Usage DTOs
// ============================

// RecordSystemMetricRequest for recording system metrics
type RecordSystemMetricRequest struct {
	MetricType    string                 `json:"metric_type" validate:"required"`
	MetricName    string                 `json:"metric_name" validate:"required"`
	MetricValue   float64                `json:"metric_value" validate:"required"`
	MetricUnit    string                 `json:"metric_unit"`
	Hour          *int                   `json:"hour,omitempty" validate:"omitempty,min=0,max=23"`
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SystemUsageMetricsResponse for system metrics responses
type SystemUsageMetricsResponse struct {
	ID                uuid.UUID              `json:"id"`
	Date              time.Time              `json:"date"`
	Hour              int                    `json:"hour"`
	MetricType        string                 `json:"metric_type"`
	MetricName        string                 `json:"metric_name"`
	MetricValue       float64                `json:"metric_value"`
	MetricUnit        string                 `json:"metric_unit"`
	APICallsTotal     int                    `json:"api_calls_total"`
	APICallsSuccess   int                    `json:"api_calls_success"`
	APICallsError     int                    `json:"api_calls_error"`
	APIResponseTime   int                    `json:"api_response_time"`
	StorageUsed       int64                  `json:"storage_used"`
	BandwidthUsed     int64                  `json:"bandwidth_used"`
	FilesUploaded     int                    `json:"files_uploaded"`
	FilesDownloaded   int                    `json:"files_downloaded"`
	DatabaseQueries   int                    `json:"database_queries"`
	DatabaseQueryTime int                    `json:"database_query_time"`
	ActiveUsers       int                    `json:"active_users"`
	NewUsers          int                    `json:"new_users"`
	SessionsCreated   int                    `json:"sessions_created"`
	LoginAttempts     int                    `json:"login_attempts"`
	LoginSuccessful   int                    `json:"login_successful"`
	OrdersCreated     int                    `json:"orders_created"`
	Revenue           float64                `json:"revenue"`
	PaymentsProcessed int                    `json:"payments_processed"`
	CustomMetrics     map[string]interface{} `json:"custom_metrics"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// SystemOverviewResponse for system overview dashboard
type SystemOverviewResponse struct {
	TenantID        string                   `json:"tenant_id"`
	Period          string                   `json:"period"`
	StartDate       time.Time                `json:"start_date"`
	EndDate         time.Time                `json:"end_date"`
	TotalAPIsCalls  int64                    `json:"total_api_calls"`
	SuccessRate     float64                  `json:"success_rate"`
	AvgResponseTime float64                  `json:"avg_response_time"`
	StorageUsed     int64                    `json:"storage_used"`
	BandwidthUsed   int64                    `json:"bandwidth_used"`
	ActiveUsers     int                      `json:"active_users"`
	TotalUsers      int                      `json:"total_users"`
	NewUsers        int                      `json:"new_users"`
	TotalSessions   int                      `json:"total_sessions"`
	Revenue         float64                  `json:"revenue"`
	OrdersCount     int                      `json:"orders_count"`
	TopMetrics      []map[string]interface{} `json:"top_metrics"`
	Trends          []map[string]interface{} `json:"trends"`
	Alerts          []map[string]interface{} `json:"alerts"`
}

// SystemMetricsListResponse for paginated system metrics
type SystemMetricsListResponse struct {
	Metrics    []*SystemUsageMetricsResponse `json:"metrics"`
	Total      int64                         `json:"total"`
	Page       int                           `json:"page"`
	Limit      int                           `json:"limit"`
	TotalPages int                           `json:"total_pages"`
}

// ============================
// Export DTOs
// ============================

// CreateReportExportRequest for creating export requests
type CreateReportExportRequest struct {
	ReportID   *uuid.UUID             `json:"report_id,omitempty"`
	ExportType string                 `json:"export_type" validate:"required,oneof=pdf excel csv"`
	ReportType string                 `json:"report_type" validate:"required"`
	Title      string                 `json:"title" validate:"required,min=3,max=255"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	DataQuery  map[string]interface{} `json:"data_query,omitempty"`
	Template   string                 `json:"template,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ReportExportResponse for export responses
type ReportExportResponse struct {
	ID             uuid.UUID  `json:"id"`
	ReportID       *uuid.UUID `json:"report_id,omitempty"`
	ExportType     string     `json:"export_type"`
	ReportType     string     `json:"report_type"`
	Title          string     `json:"title"`
	FileURL        string     `json:"file_url,omitempty"`
	FileSize       int64      `json:"file_size"`
	Status         string     `json:"status"`
	Progress       int        `json:"progress"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	RowsProcessed  int        `json:"rows_processed"`
	TotalRows      int        `json:"total_rows"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	DownloadCount  int        `json:"download_count"`
	LastDownloadAt *time.Time `json:"last_download_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ExportListResponse for paginated export lists
type ExportListResponse struct {
	Exports    []*ReportExportResponse `json:"exports"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
	TotalPages int                     `json:"total_pages"`
}

// ============================
// Report Schedule DTOs
// ============================

// CreateReportScheduleRequest for creating scheduled reports
type CreateReportScheduleRequest struct {
	Name            string                 `json:"name" validate:"required,min=3,max=255"`
	Description     string                 `json:"description,omitempty"`
	ReportType      string                 `json:"report_type" validate:"required"`
	ReportSubtype   string                 `json:"report_subtype,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	Schedule        string                 `json:"schedule" validate:"required"` // cron expression
	OutputFormats   []string               `json:"output_formats" validate:"required,dive,oneof=pdf excel csv json"`
	EmailRecipients []string               `json:"email_recipients,omitempty" validate:"dive,email"`
	RetentionDays   int                    `json:"retention_days" validate:"min=1,max=365"`
	IsActive        bool                   `json:"is_active"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateReportScheduleRequest for updating schedules
type UpdateReportScheduleRequest struct {
	Name            *string                `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description     *string                `json:"description,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	Schedule        *string                `json:"schedule,omitempty"`
	OutputFormats   []string               `json:"output_formats,omitempty" validate:"omitempty,dive,oneof=pdf excel csv json"`
	EmailRecipients []string               `json:"email_recipients,omitempty" validate:"omitempty,dive,email"`
	RetentionDays   *int                   `json:"retention_days,omitempty" validate:"omitempty,min=1,max=365"`
	IsActive        *bool                  `json:"is_active,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ReportScheduleResponse for schedule responses
type ReportScheduleResponse struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	ReportType      string                 `json:"report_type"`
	ReportSubtype   string                 `json:"report_subtype"`
	Parameters      map[string]interface{} `json:"parameters"`
	Schedule        string                 `json:"schedule"`
	OutputFormats   []string               `json:"output_formats"`
	EmailRecipients []string               `json:"email_recipients"`
	RetentionDays   int                    `json:"retention_days"`
	IsActive        bool                   `json:"is_active"`
	LastRunAt       *time.Time             `json:"last_run_at,omitempty"`
	NextRunAt       *time.Time             `json:"next_run_at,omitempty"`
	RunCount        int                    `json:"run_count"`
	ErrorCount      int                    `json:"error_count"`
	LastError       string                 `json:"last_error,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ScheduleListResponse for paginated schedule lists
type ScheduleListResponse struct {
	Schedules  []*ReportScheduleResponse `json:"schedules"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	Limit      int                       `json:"limit"`
	TotalPages int                       `json:"total_pages"`
}

// ============================
// Dashboard DTOs
// ============================

// DashboardStatsResponse for dashboard statistics
type DashboardStatsResponse struct {
	TenantID      string                     `json:"tenant_id"`
	Period        string                     `json:"period"`
	UserStats     map[string]interface{}     `json:"user_stats"`
	SystemStats   map[string]interface{}     `json:"system_stats"`
	SalesStats    map[string]interface{}     `json:"sales_stats,omitempty"`
	RecentReports []*AnalyticsReportResponse `json:"recent_reports"`
	RecentExports []*ReportExportResponse    `json:"recent_exports"`
	AlertsCount   int                        `json:"alerts_count"`
	QuickMetrics  []map[string]interface{}   `json:"quick_metrics"`
	TrendData     []map[string]interface{}   `json:"trend_data"`
	GeneratedAt   time.Time                  `json:"generated_at"`
}
