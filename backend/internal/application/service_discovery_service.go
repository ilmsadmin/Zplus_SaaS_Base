package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// ServiceDiscoveryService handles service registration and health monitoring
type ServiceDiscoveryService struct {
	serviceRepo     domain.FederationServiceRepository
	schemaRepo      domain.GraphQLSchemaRepository
	changeEventRepo domain.SchemaChangeEventRepository
	logger          *zap.Logger
	httpClient      *http.Client
}

// NewServiceDiscoveryService creates a new service discovery service
func NewServiceDiscoveryService(
	serviceRepo domain.FederationServiceRepository,
	schemaRepo domain.GraphQLSchemaRepository,
	changeEventRepo domain.SchemaChangeEventRepository,
	logger *zap.Logger,
) *ServiceDiscoveryService {
	return &ServiceDiscoveryService{
		serviceRepo:     serviceRepo,
		schemaRepo:      schemaRepo,
		changeEventRepo: changeEventRepo,
		logger:          logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterServiceRequest represents a service registration request
type RegisterServiceRequest struct {
	ServiceName    string                 `json:"service_name" validate:"required"`
	ServiceURL     string                 `json:"service_url" validate:"required"`
	HealthCheckURL string                 `json:"health_check_url"`
	Tags           []string               `json:"tags"`
	Weight         int                    `json:"weight"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// RegisterServiceResponse represents a service registration response
type RegisterServiceResponse struct {
	ServiceID    string `json:"service_id"`
	ServiceName  string `json:"service_name"`
	Status       string `json:"status"`
	RegisteredAt string `json:"registered_at"`
}

// ServiceHealth represents service health check result
type ServiceHealth struct {
	ServiceName  string            `json:"service_name"`
	Status       string            `json:"status"`
	ResponseTime time.Duration     `json:"response_time"`
	LastCheck    time.Time         `json:"last_check"`
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]string `json:"metadata"`
}

// RegisterService registers a new service for discovery
func (s *ServiceDiscoveryService) RegisterService(ctx context.Context, req *RegisterServiceRequest) (*RegisterServiceResponse, error) {
	s.logger.Info("Registering service for discovery",
		zap.String("service", req.ServiceName),
		zap.String("url", req.ServiceURL),
	)

	// Set defaults
	if req.HealthCheckURL == "" {
		req.HealthCheckURL = req.ServiceURL + "/health"
	}
	if req.Weight <= 0 {
		req.Weight = 100
	}
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	if req.Tags == nil {
		req.Tags = []string{"graphql", "federation"}
	}

	// Create service record
	service := &domain.FederationService{
		ServiceName:    req.ServiceName,
		ServiceURL:     req.ServiceURL,
		HealthCheckURL: req.HealthCheckURL,
		Status:         domain.ServiceStatusUnknown,
		Tags:           req.Tags,
		Weight:         req.Weight,
		Metadata:       req.Metadata,
	}

	// Check if service already exists
	existingService, err := s.serviceRepo.GetByName(ctx, req.ServiceName)
	if err == nil && existingService != nil {
		// Update existing service
		existingService.ServiceURL = req.ServiceURL
		existingService.HealthCheckURL = req.HealthCheckURL
		existingService.Tags = req.Tags
		existingService.Weight = req.Weight
		existingService.Metadata = req.Metadata

		if err := s.serviceRepo.Update(ctx, existingService); err != nil {
			s.logger.Error("Failed to update existing service",
				zap.Error(err),
				zap.String("service", req.ServiceName),
			)
			return nil, fmt.Errorf("failed to update service: %w", err)
		}

		service = existingService
	} else {
		// Register new service
		if err := s.serviceRepo.Register(ctx, service); err != nil {
			s.logger.Error("Failed to register service",
				zap.Error(err),
				zap.String("service", req.ServiceName),
			)
			return nil, fmt.Errorf("failed to register service: %w", err)
		}

		// Create registration event
		changeEvent := &domain.SchemaChangeEvent{
			ServiceName: req.ServiceName,
			ChangeType:  domain.ChangeTypeServiceRegistered,
			ChangeDetails: map[string]interface{}{
				"service_url":      req.ServiceURL,
				"health_check_url": req.HealthCheckURL,
				"tags":             req.Tags,
				"weight":           req.Weight,
			},
		}

		if err := s.changeEventRepo.Create(ctx, changeEvent); err != nil {
			s.logger.Error("Failed to create registration event",
				zap.Error(err),
				zap.String("service", req.ServiceName),
			)
		}
	}

	// Perform initial health check
	go s.performHealthCheck(context.Background(), service)

	s.logger.Info("Service registration completed",
		zap.String("service", req.ServiceName),
		zap.String("service_id", service.ID.String()),
	)

	return &RegisterServiceResponse{
		ServiceID:    service.ID.String(),
		ServiceName:  service.ServiceName,
		Status:       service.Status,
		RegisteredAt: service.CreatedAt.Format(time.RFC3339),
	}, nil
}

// DeregisterService removes a service from discovery
func (s *ServiceDiscoveryService) DeregisterService(ctx context.Context, serviceName string) error {
	s.logger.Info("Deregistering service", zap.String("service", serviceName))

	service, err := s.serviceRepo.GetByName(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// Create deregistration event
	changeEvent := &domain.SchemaChangeEvent{
		ServiceName: serviceName,
		ChangeType:  domain.ChangeTypeServiceDeregistered,
		ChangeDetails: map[string]interface{}{
			"service_url": service.ServiceURL,
			"last_status": service.Status,
		},
	}

	if err := s.changeEventRepo.Create(ctx, changeEvent); err != nil {
		s.logger.Error("Failed to create deregistration event",
			zap.Error(err),
			zap.String("service", serviceName),
		)
	}

	return s.serviceRepo.Deregister(ctx, serviceName)
}

// GetHealthyServices returns all healthy services
func (s *ServiceDiscoveryService) GetHealthyServices(ctx context.Context) ([]*domain.FederationService, error) {
	return s.serviceRepo.GetHealthyServices(ctx)
}

// GetAllServices returns all registered services
func (s *ServiceDiscoveryService) GetAllServices(ctx context.Context) ([]*domain.FederationService, error) {
	return s.serviceRepo.ListActive(ctx)
}

// GetServiceByName returns a specific service by name
func (s *ServiceDiscoveryService) GetServiceByName(ctx context.Context, serviceName string) (*domain.FederationService, error) {
	return s.serviceRepo.GetByName(ctx, serviceName)
}

// PerformHealthChecks performs health checks on all registered services
func (s *ServiceDiscoveryService) PerformHealthChecks(ctx context.Context) ([]ServiceHealth, error) {
	services, err := s.serviceRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	var healthResults []ServiceHealth

	for _, service := range services {
		health := s.performHealthCheck(ctx, service)
		healthResults = append(healthResults, health)
	}

	return healthResults, nil
}

// StartHealthMonitoring starts periodic health monitoring
func (s *ServiceDiscoveryService) StartHealthMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("Starting health monitoring", zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Health monitoring stopped")
			return
		case <-ticker.C:
			s.performAllHealthChecks(ctx)
		}
	}
}

// performAllHealthChecks performs health checks on all services
func (s *ServiceDiscoveryService) performAllHealthChecks(ctx context.Context) {
	services, err := s.serviceRepo.ListActive(ctx)
	if err != nil {
		s.logger.Error("Failed to get services for health check", zap.Error(err))
		return
	}

	for _, service := range services {
		go s.performHealthCheck(ctx, service)
	}
}

// performHealthCheck performs a health check on a single service
func (s *ServiceDiscoveryService) performHealthCheck(ctx context.Context, service *domain.FederationService) ServiceHealth {
	start := time.Now()

	health := ServiceHealth{
		ServiceName: service.ServiceName,
		LastCheck:   start,
		Metadata:    make(map[string]string),
	}

	// Perform HTTP health check
	req, err := http.NewRequestWithContext(ctx, "GET", service.HealthCheckURL, nil)
	if err != nil {
		health.Status = domain.ServiceStatusUnhealthy
		health.Error = fmt.Sprintf("Failed to create request: %v", err)
		s.updateServiceStatus(ctx, service.ServiceName, health.Status)
		return health
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		health.Status = domain.ServiceStatusUnhealthy
		health.Error = fmt.Sprintf("Health check failed: %v", err)
		health.ResponseTime = time.Since(start)
		s.updateServiceStatus(ctx, service.ServiceName, health.Status)
		return health
	}
	defer resp.Body.Close()

	health.ResponseTime = time.Since(start)

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		health.Status = domain.ServiceStatusHealthy
		health.Metadata["status_code"] = fmt.Sprintf("%d", resp.StatusCode)
	} else {
		health.Status = domain.ServiceStatusUnhealthy
		health.Error = fmt.Sprintf("Unhealthy status code: %d", resp.StatusCode)
		health.Metadata["status_code"] = fmt.Sprintf("%d", resp.StatusCode)
	}

	// Update service status in database
	s.updateServiceStatus(ctx, service.ServiceName, health.Status)

	s.logger.Debug("Health check completed",
		zap.String("service", service.ServiceName),
		zap.String("status", health.Status),
		zap.Duration("response_time", health.ResponseTime),
	)

	return health
}

// updateServiceStatus updates the service status in the database
func (s *ServiceDiscoveryService) updateServiceStatus(ctx context.Context, serviceName, status string) {
	if err := s.serviceRepo.UpdateStatus(ctx, serviceName, status); err != nil {
		s.logger.Error("Failed to update service status",
			zap.Error(err),
			zap.String("service", serviceName),
			zap.String("status", status),
		)
	}

	if err := s.serviceRepo.UpdateHealthCheck(ctx, serviceName, time.Now()); err != nil {
		s.logger.Error("Failed to update health check timestamp",
			zap.Error(err),
			zap.String("service", serviceName),
		)
	}
}
