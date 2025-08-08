package services

import (
	"time"

	"github.com/google/uuid"
)

// ===========================
// Tenant Service DTOs
// ===========================

// CreateTenantRequest represents request to create a new tenant
type CreateTenantRequest struct {
	Name         string   `json:"name" validate:"required,min=2,max=100"`
	Subdomain    string   `json:"subdomain" validate:"required,min=3,max=50,alphanum"`
	Plan         string   `json:"plan" validate:"required"`
	ContactEmail string   `json:"contact_email" validate:"required,email"`
	ContactName  string   `json:"contact_name" validate:"required"`
	ContactPhone string   `json:"contact_phone"`
	CompanySize  string   `json:"company_size"`
	Industry     string   `json:"industry"`
	Region       string   `json:"region"`
	Timezone     string   `json:"timezone"`
	Language     string   `json:"language"`
	Currency     string   `json:"currency"`
	Features     []string `json:"features"`
}

// UpdateTenantRequest represents request to update a tenant
type UpdateTenantRequest struct {
	Name         *string  `json:"name" validate:"omitempty,min=2,max=100"`
	Status       *string  `json:"status" validate:"omitempty,oneof=pending active suspended cancelled"`
	Plan         *string  `json:"plan" validate:"omitempty"`
	ContactEmail *string  `json:"contact_email" validate:"omitempty,email"`
	ContactName  *string  `json:"contact_name"`
	ContactPhone *string  `json:"contact_phone"`
	CompanySize  *string  `json:"company_size"`
	Industry     *string  `json:"industry"`
	Region       *string  `json:"region"`
	Timezone     *string  `json:"timezone"`
	Language     *string  `json:"language"`
	Currency     *string  `json:"currency"`
	Features     []string `json:"features"`
}

// TenantResponse represents tenant data in responses
type TenantResponse struct {
	ID                 uuid.UUID              `json:"id"`
	Name               string                 `json:"name"`
	Subdomain          string                 `json:"subdomain"`
	Status             string                 `json:"status"`
	Plan               string                 `json:"plan"`
	OnboardingStatus   string                 `json:"onboarding_status"`
	OnboardingStep     int                    `json:"onboarding_step"`
	ContactEmail       string                 `json:"contact_email"`
	ContactName        string                 `json:"contact_name"`
	ContactPhone       string                 `json:"contact_phone"`
	CompanySize        string                 `json:"company_size"`
	Industry           string                 `json:"industry"`
	Region             string                 `json:"region"`
	Timezone           string                 `json:"timezone"`
	Language           string                 `json:"language"`
	Currency           string                 `json:"currency"`
	Features           []string               `json:"features"`
	TrialStartsAt      *time.Time             `json:"trial_starts_at"`
	TrialEndsAt        *time.Time             `json:"trial_ends_at"`
	SubscriptionStatus string                 `json:"subscription_status"`
	LastActivityAt     *time.Time             `json:"last_activity_at"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	Settings           map[string]interface{} `json:"settings,omitempty"`
	Configuration      map[string]interface{} `json:"configuration,omitempty"`
	Branding           map[string]interface{} `json:"branding,omitempty"`
	UserCount          int                    `json:"user_count,omitempty"`
	DomainCount        int                    `json:"domain_count,omitempty"`
}

// ListTenantsRequest represents request to list tenants
type ListTenantsRequest struct {
	Page       int    `json:"page" validate:"min=1"`
	Limit      int    `json:"limit" validate:"min=1,max=100"`
	Sort       string `json:"sort"`
	Filter     string `json:"filter"`
	Status     string `json:"status"`
	Plan       string `json:"plan"`
	Search     string `json:"search"`
	OnlyActive bool   `json:"only_active"`
	OnlyTrial  bool   `json:"only_trial"`
}

// TenantListResponse represents response for tenant list
type TenantListResponse struct {
	Tenants    []*TenantResponse `json:"tenants"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

// ===========================
// Tenant Onboarding DTOs
// ===========================

// StartOnboardingRequest represents request to start tenant onboarding
type StartOnboardingRequest struct {
	TenantID uuid.UUID              `json:"tenant_id" validate:"required"`
	Data     map[string]interface{} `json:"data"`
}

// CompleteOnboardingStepRequest represents request to complete an onboarding step
type CompleteOnboardingStepRequest struct {
	TenantID uuid.UUID              `json:"tenant_id" validate:"required"`
	Step     int                    `json:"step" validate:"required,min=1"`
	Data     map[string]interface{} `json:"data"`
}

// OnboardingStepResponse represents onboarding step data
type OnboardingStepResponse struct {
	Step        int                    `json:"step"`
	StepName    string                 `json:"step_name"`
	Status      string                 `json:"status"`
	Data        map[string]interface{} `json:"data"`
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
}

// OnboardingStatusResponse represents overall onboarding status
type OnboardingStatusResponse struct {
	TenantID          uuid.UUID                 `json:"tenant_id"`
	Status            string                    `json:"status"`
	CurrentStep       int                       `json:"current_step"`
	TotalSteps        int                       `json:"total_steps"`
	CompletedSteps    int                       `json:"completed_steps"`
	ProgressPercent   float64                   `json:"progress_percent"`
	Steps             []*OnboardingStepResponse `json:"steps"`
	EstimatedTimeLeft string                    `json:"estimated_time_left,omitempty"`
}

// ===========================
// Tenant Configuration DTOs
// ===========================

// UpdateTenantSettingsRequest represents request to update tenant settings
type UpdateTenantSettingsRequest struct {
	Theme                string                 `json:"theme"`
	Language             string                 `json:"language"`
	Timezone             string                 `json:"timezone"`
	DateFormat           string                 `json:"date_format"`
	TimeFormat           string                 `json:"time_format"`
	NumberFormat         string                 `json:"number_format"`
	Features             []string               `json:"features"`
	APILimits            map[string]interface{} `json:"api_limits"`
	SecuritySettings     map[string]interface{} `json:"security_settings"`
	NotificationSettings map[string]interface{} `json:"notification_settings"`
	IntegrationSettings  map[string]interface{} `json:"integration_settings"`
}

// UpdateTenantConfigurationRequest represents request to update tenant configuration
type UpdateTenantConfigurationRequest struct {
	MaxUsers            *int                   `json:"max_users"`
	MaxStorage          *int64                 `json:"max_storage"`
	MaxAPICallsPerMonth *int                   `json:"max_api_calls_per_month"`
	EnabledModules      []string               `json:"enabled_modules"`
	CustomDomainEnabled *bool                  `json:"custom_domain_enabled"`
	SSLEnabled          *bool                  `json:"ssl_enabled"`
	BackupEnabled       *bool                  `json:"backup_enabled"`
	MonitoringEnabled   *bool                  `json:"monitoring_enabled"`
	AuditLoggingEnabled *bool                  `json:"audit_logging_enabled"`
	DataRetentionDays   *int                   `json:"data_retention_days"`
	AllowedFileTypes    []string               `json:"allowed_file_types"`
	MaxFileSize         *int64                 `json:"max_file_size"`
	NetworkSettings     map[string]interface{} `json:"network_settings"`
	DatabaseSettings    map[string]interface{} `json:"database_settings"`
	CacheSettings       map[string]interface{} `json:"cache_settings"`
}

// UpdateTenantBrandingRequest represents request to update tenant branding
type UpdateTenantBrandingRequest struct {
	Logo              *string                `json:"logo"`
	LogoURL           *string                `json:"logo_url"`
	FaviconURL        *string                `json:"favicon_url"`
	PrimaryColor      *string                `json:"primary_color"`
	SecondaryColor    *string                `json:"secondary_color"`
	AccentColor       *string                `json:"accent_color"`
	BackgroundColor   *string                `json:"background_color"`
	TextColor         *string                `json:"text_color"`
	FontFamily        *string                `json:"font_family"`
	CustomCSS         *string                `json:"custom_css"`
	CustomJavaScript  *string                `json:"custom_javascript"`
	FooterText        *string                `json:"footer_text"`
	TermsOfServiceURL *string                `json:"terms_of_service_url"`
	PrivacyPolicyURL  *string                `json:"privacy_policy_url"`
	SupportURL        *string                `json:"support_url"`
	DocumentationURL  *string                `json:"documentation_url"`
	CustomDomainName  *string                `json:"custom_domain_name"`
	EmailFromName     *string                `json:"email_from_name"`
	EmailFromAddress  *string                `json:"email_from_address"`
	SocialMediaLinks  map[string]interface{} `json:"social_media_links"`
	CustomMetaTags    map[string]interface{} `json:"custom_meta_tags"`
	PWASettings       map[string]interface{} `json:"pwa_settings"`
}

// ===========================
// Tenant Invitation DTOs
// ===========================

// CreateTenantInvitationRequest represents request to create tenant invitation
type CreateTenantInvitationRequest struct {
	TenantID uuid.UUID              `json:"tenant_id" validate:"required"`
	Email    string                 `json:"email" validate:"required,email"`
	Role     string                 `json:"role" validate:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}

// TenantInvitationResponse represents tenant invitation data
type TenantInvitationResponse struct {
	ID         uuid.UUID              `json:"id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	Email      string                 `json:"email"`
	Role       string                 `json:"role"`
	Status     string                 `json:"status"`
	ExpiresAt  time.Time              `json:"expires_at"`
	AcceptedAt *time.Time             `json:"accepted_at"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// AcceptInvitationRequest represents request to accept tenant invitation
type AcceptInvitationRequest struct {
	Token  string    `json:"token" validate:"required"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// ===========================
// Subdomain Management DTOs
// ===========================

// CheckSubdomainAvailabilityRequest represents request to check subdomain availability
type CheckSubdomainAvailabilityRequest struct {
	Subdomain string `json:"subdomain" validate:"required,min=3,max=50,alphanum"`
}

// SubdomainAvailabilityResponse represents subdomain availability result
type SubdomainAvailabilityResponse struct {
	Subdomain   string   `json:"subdomain"`
	Available   bool     `json:"available"`
	Suggestions []string `json:"suggestions,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

// ===========================
// Tenant Metrics DTOs
// ===========================

// RecordUsageMetricRequest represents request to record usage metric
type RecordUsageMetricRequest struct {
	TenantID       uuid.UUID              `json:"tenant_id" validate:"required"`
	MetricType     string                 `json:"metric_type" validate:"required"`
	Value          float64                `json:"value" validate:"required"`
	Unit           string                 `json:"unit" validate:"required"`
	AdditionalData map[string]interface{} `json:"additional_data"`
}

// TenantMetricsResponse represents tenant metrics data
type TenantMetricsResponse struct {
	TenantID    uuid.UUID              `json:"tenant_id"`
	Metrics     map[string]interface{} `json:"metrics"`
	Period      map[string]interface{} `json:"period"`
	Summary     map[string]interface{} `json:"summary"`
	LastUpdated time.Time              `json:"last_updated"`
}

// GetTenantMetricsRequest represents request to get tenant metrics
type GetTenantMetricsRequest struct {
	TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
	MetricTypes []string  `json:"metric_types"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	Granularity string    `json:"granularity"` // hour, day, week, month
}

// TenantStatsResponse represents tenant statistics
type TenantStatsResponse struct {
	TenantID          uuid.UUID              `json:"tenant_id"`
	UserCount         int                    `json:"user_count"`
	ActiveUserCount   int                    `json:"active_user_count"`
	DomainCount       int                    `json:"domain_count"`
	StorageUsed       int64                  `json:"storage_used"`
	APICallsThisMonth int                    `json:"api_calls_this_month"`
	LastActivityAt    *time.Time             `json:"last_activity_at"`
	UploadedFiles     int                    `json:"uploaded_files"`
	CustomData        map[string]interface{} `json:"custom_data"`
}

// ===========================
// Billing Integration DTOs
// ===========================

// UpdateTenantBillingRequest represents request to update tenant billing
type UpdateTenantBillingRequest struct {
	StripeCustomerID     *string                `json:"stripe_customer_id"`
	BillingEmail         *string                `json:"billing_email" validate:"omitempty,email"`
	BillingName          *string                `json:"billing_name"`
	BillingAddress       map[string]interface{} `json:"billing_address"`
	TaxInfo              map[string]interface{} `json:"tax_info"`
	PaymentMethodID      *string                `json:"payment_method_id"`
	DefaultPaymentMethod *string                `json:"default_payment_method"`
	BillingCycle         *string                `json:"billing_cycle"`
	AutoBilling          *bool                  `json:"auto_billing"`
	ProrationBehavior    *string                `json:"proration_behavior"`
}

// TenantBillingResponse represents tenant billing information
type TenantBillingResponse struct {
	TenantID           uuid.UUID              `json:"tenant_id"`
	StripeCustomerID   string                 `json:"stripe_customer_id,omitempty"`
	BillingEmail       string                 `json:"billing_email"`
	BillingName        string                 `json:"billing_name"`
	BillingAddress     map[string]interface{} `json:"billing_address"`
	TaxInfo            map[string]interface{} `json:"tax_info"`
	BillingCycle       string                 `json:"billing_cycle"`
	NextBillingDate    *time.Time             `json:"next_billing_date"`
	LastBillingDate    *time.Time             `json:"last_billing_date"`
	CurrentPeriodStart *time.Time             `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time             `json:"current_period_end"`
	TrialPeriodDays    int                    `json:"trial_period_days"`
	CreditBalance      float64                `json:"credit_balance"`
	AutoBilling        bool                   `json:"auto_billing"`
}
