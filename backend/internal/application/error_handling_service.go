package application

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ErrorHandlingService handles standardized error management for GraphQL Federation
type ErrorHandlingService struct {
	logger *zap.Logger
}

// NewErrorHandlingService creates a new error handling service
func NewErrorHandlingService(logger *zap.Logger) *ErrorHandlingService {
	return &ErrorHandlingService{
		logger: logger,
	}
}

// FederationError represents a standardized federation error
type FederationError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Severity  string                 `json:"severity"`
	Retryable bool                   `json:"retryable"`
	Category  string                 `json:"category"`
}

// Error codes
const (
	ErrCodeSchemaValidation     = "SCHEMA_VALIDATION_ERROR"
	ErrCodeServiceUnavailable   = "SERVICE_UNAVAILABLE"
	ErrCodeCompositionFailed    = "COMPOSITION_FAILED"
	ErrCodeQueryPlanningFailed  = "QUERY_PLANNING_FAILED"
	ErrCodeExecutionTimeout     = "EXECUTION_TIMEOUT"
	ErrCodeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"
	ErrCodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	ErrCodeAuthorizationFailed  = "AUTHORIZATION_FAILED"
	ErrCodeInternalError        = "INTERNAL_ERROR"
	ErrCodeConfigurationError   = "CONFIGURATION_ERROR"
	ErrCodeNetworkError         = "NETWORK_ERROR"
	ErrCodeDataIntegrityError   = "DATA_INTEGRITY_ERROR"
)

// Error severities
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// Error categories
const (
	CategoryValidation     = "validation"
	CategoryNetwork        = "network"
	CategoryAuthentication = "authentication"
	CategoryAuthorization  = "authorization"
	CategoryConfiguration  = "configuration"
	CategoryExecution      = "execution"
	CategorySystem         = "system"
)

// CreateError creates a standardized federation error
func (e *ErrorHandlingService) CreateError(code, message, service, operation string, details map[string]interface{}) *FederationError {
	severity := e.getSeverityForCode(code)
	category := e.getCategoryForCode(code)
	retryable := e.isRetryableError(code)

	return &FederationError{
		Code:      code,
		Message:   message,
		Service:   service,
		Operation: operation,
		Details:   details,
		Timestamp: time.Now(),
		Severity:  severity,
		Retryable: retryable,
		Category:  category,
	}
}

// HandleSchemaValidationError handles schema validation errors
func (e *ErrorHandlingService) HandleSchemaValidationError(ctx context.Context, serviceName string, validationErrors []string) *FederationError {
	details := map[string]interface{}{
		"validation_errors": validationErrors,
		"error_count":       len(validationErrors),
	}

	e.logger.Error("Schema validation failed",
		zap.String("service", serviceName),
		zap.Strings("validation_errors", validationErrors),
	)

	return e.CreateError(
		ErrCodeSchemaValidation,
		fmt.Sprintf("Schema validation failed for service %s", serviceName),
		serviceName,
		"schema_validation",
		details,
	)
}

// HandleServiceUnavailableError handles service unavailable errors
func (e *ErrorHandlingService) HandleServiceUnavailableError(ctx context.Context, serviceName, serviceURL string, err error) *FederationError {
	details := map[string]interface{}{
		"service_url": serviceURL,
		"error":       err.Error(),
	}

	e.logger.Error("Service unavailable",
		zap.String("service", serviceName),
		zap.String("url", serviceURL),
		zap.Error(err),
	)

	return e.CreateError(
		ErrCodeServiceUnavailable,
		fmt.Sprintf("Service %s is unavailable", serviceName),
		serviceName,
		"health_check",
		details,
	)
}

// HandleCompositionError handles schema composition errors
func (e *ErrorHandlingService) HandleCompositionError(ctx context.Context, services []string, err error) *FederationError {
	details := map[string]interface{}{
		"services":      services,
		"service_count": len(services),
		"error":         err.Error(),
	}

	e.logger.Error("Schema composition failed",
		zap.Strings("services", services),
		zap.Error(err),
	)

	return e.CreateError(
		ErrCodeCompositionFailed,
		"Failed to compose federated schema",
		"gateway",
		"schema_composition",
		details,
	)
}

// HandleQueryExecutionError handles query execution errors
func (e *ErrorHandlingService) HandleQueryExecutionError(ctx context.Context, queryID string, err error, executionTime time.Duration) *FederationError {
	details := map[string]interface{}{
		"query_id":       queryID,
		"execution_time": executionTime.Milliseconds(),
		"error":          err.Error(),
	}

	errorCode := ErrCodeInternalError
	if executionTime > 30*time.Second {
		errorCode = ErrCodeExecutionTimeout
	}

	e.logger.Error("Query execution failed",
		zap.String("query_id", queryID),
		zap.Duration("execution_time", executionTime),
		zap.Error(err),
	)

	return e.CreateError(
		errorCode,
		"Query execution failed",
		"gateway",
		"query_execution",
		details,
	)
}

// HandleAuthenticationError handles authentication errors
func (e *ErrorHandlingService) HandleAuthenticationError(ctx context.Context, reason string, details map[string]interface{}) *FederationError {
	e.logger.Warn("Authentication failed",
		zap.String("reason", reason),
		zap.Any("details", details),
	)

	return e.CreateError(
		ErrCodeAuthenticationFailed,
		fmt.Sprintf("Authentication failed: %s", reason),
		"gateway",
		"authentication",
		details,
	)
}

// HandleAuthorizationError handles authorization errors
func (e *ErrorHandlingService) HandleAuthorizationError(ctx context.Context, operation, resource string, userID string) *FederationError {
	details := map[string]interface{}{
		"operation": operation,
		"resource":  resource,
		"user_id":   userID,
	}

	e.logger.Warn("Authorization failed",
		zap.String("operation", operation),
		zap.String("resource", resource),
		zap.String("user_id", userID),
	)

	return e.CreateError(
		ErrCodeAuthorizationFailed,
		fmt.Sprintf("Insufficient permissions for operation %s on resource %s", operation, resource),
		"gateway",
		"authorization",
		details,
	)
}

// HandleRateLimitError handles rate limiting errors
func (e *ErrorHandlingService) HandleRateLimitError(ctx context.Context, identifier string, limit int, window time.Duration) *FederationError {
	details := map[string]interface{}{
		"identifier": identifier,
		"limit":      limit,
		"window_ms":  window.Milliseconds(),
	}

	e.logger.Warn("Rate limit exceeded",
		zap.String("identifier", identifier),
		zap.Int("limit", limit),
		zap.Duration("window", window),
	)

	return e.CreateError(
		ErrCodeRateLimitExceeded,
		fmt.Sprintf("Rate limit of %d requests per %v exceeded", limit, window),
		"gateway",
		"rate_limiting",
		details,
	)
}

// HandleNetworkError handles network-related errors
func (e *ErrorHandlingService) HandleNetworkError(ctx context.Context, serviceName, operation string, err error) *FederationError {
	details := map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
	}

	e.logger.Error("Network error",
		zap.String("service", serviceName),
		zap.String("operation", operation),
		zap.Error(err),
	)

	return e.CreateError(
		ErrCodeNetworkError,
		fmt.Sprintf("Network error during %s", operation),
		serviceName,
		operation,
		details,
	)
}

// HandleConfigurationError handles configuration errors
func (e *ErrorHandlingService) HandleConfigurationError(ctx context.Context, configKey string, err error) *FederationError {
	details := map[string]interface{}{
		"config_key": configKey,
		"error":      err.Error(),
	}

	e.logger.Error("Configuration error",
		zap.String("config_key", configKey),
		zap.Error(err),
	)

	return e.CreateError(
		ErrCodeConfigurationError,
		fmt.Sprintf("Configuration error for key %s", configKey),
		"gateway",
		"configuration",
		details,
	)
}

// IsRetryable checks if an error is retryable
func (e *ErrorHandlingService) IsRetryable(err *FederationError) bool {
	return err.Retryable
}

// GetRetryDelay calculates retry delay based on attempt count
func (e *ErrorHandlingService) GetRetryDelay(attempt int, baseDelay time.Duration) time.Duration {
	// Exponential backoff with jitter
	delay := baseDelay * time.Duration(1<<uint(attempt))
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	return delay
}

// LogError logs an error with appropriate level
func (e *ErrorHandlingService) LogError(err *FederationError) {
	fields := []zap.Field{
		zap.String("code", err.Code),
		zap.String("service", err.Service),
		zap.String("operation", err.Operation),
		zap.String("category", err.Category),
		zap.Bool("retryable", err.Retryable),
		zap.Any("details", err.Details),
	}

	switch err.Severity {
	case SeverityCritical:
		e.logger.Error(err.Message, fields...)
	case SeverityHigh:
		e.logger.Error(err.Message, fields...)
	case SeverityMedium:
		e.logger.Warn(err.Message, fields...)
	case SeverityLow:
		e.logger.Info(err.Message, fields...)
	case SeverityInfo:
		e.logger.Info(err.Message, fields...)
	default:
		e.logger.Error(err.Message, fields...)
	}
}

// getSeverityForCode returns the severity level for an error code
func (e *ErrorHandlingService) getSeverityForCode(code string) string {
	switch code {
	case ErrCodeInternalError, ErrCodeDataIntegrityError:
		return SeverityCritical
	case ErrCodeServiceUnavailable, ErrCodeCompositionFailed, ErrCodeConfigurationError:
		return SeverityHigh
	case ErrCodeQueryPlanningFailed, ErrCodeNetworkError, ErrCodeExecutionTimeout:
		return SeverityMedium
	case ErrCodeSchemaValidation, ErrCodeAuthenticationFailed, ErrCodeAuthorizationFailed:
		return SeverityLow
	case ErrCodeRateLimitExceeded:
		return SeverityInfo
	default:
		return SeverityMedium
	}
}

// getCategoryForCode returns the category for an error code
func (e *ErrorHandlingService) getCategoryForCode(code string) string {
	switch code {
	case ErrCodeSchemaValidation:
		return CategoryValidation
	case ErrCodeServiceUnavailable, ErrCodeNetworkError:
		return CategoryNetwork
	case ErrCodeAuthenticationFailed:
		return CategoryAuthentication
	case ErrCodeAuthorizationFailed:
		return CategoryAuthorization
	case ErrCodeConfigurationError:
		return CategoryConfiguration
	case ErrCodeCompositionFailed, ErrCodeQueryPlanningFailed, ErrCodeExecutionTimeout:
		return CategoryExecution
	default:
		return CategorySystem
	}
}

// isRetryableError determines if an error code represents a retryable error
func (e *ErrorHandlingService) isRetryableError(code string) bool {
	switch code {
	case ErrCodeServiceUnavailable, ErrCodeNetworkError, ErrCodeExecutionTimeout:
		return true
	case ErrCodeSchemaValidation, ErrCodeAuthenticationFailed, ErrCodeAuthorizationFailed, ErrCodeConfigurationError:
		return false
	default:
		return false
	}
}
