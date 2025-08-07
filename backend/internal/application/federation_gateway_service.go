package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// FederationGatewayService handles GraphQL federation and query routing
type FederationGatewayService struct {
	serviceRepo       domain.FederationServiceRepository
	schemaRepo        domain.GraphQLSchemaRepository
	compositionRepo   domain.FederationCompositionRepository
	gatewayConfigRepo domain.FederationGatewayConfigRepository
	queryMetricsRepo  domain.GraphQLQueryMetricsRepository
	changeEventRepo   domain.SchemaChangeEventRepository
	serviceDiscovery  *ServiceDiscoveryService
	logger            *zap.Logger

	// Cache for compositions and configurations
	compositionCache sync.Map
	configCache      sync.Map
	lastCacheUpdate  time.Time
	cacheMutex       sync.RWMutex
}

// NewFederationGatewayService creates a new federation gateway service
func NewFederationGatewayService(
	serviceRepo domain.FederationServiceRepository,
	schemaRepo domain.GraphQLSchemaRepository,
	compositionRepo domain.FederationCompositionRepository,
	gatewayConfigRepo domain.FederationGatewayConfigRepository,
	queryMetricsRepo domain.GraphQLQueryMetricsRepository,
	changeEventRepo domain.SchemaChangeEventRepository,
	serviceDiscovery *ServiceDiscoveryService,
	logger *zap.Logger,
) *FederationGatewayService {
	return &FederationGatewayService{
		serviceRepo:       serviceRepo,
		schemaRepo:        schemaRepo,
		compositionRepo:   compositionRepo,
		gatewayConfigRepo: gatewayConfigRepo,
		queryMetricsRepo:  queryMetricsRepo,
		changeEventRepo:   changeEventRepo,
		serviceDiscovery:  serviceDiscovery,
		logger:            logger,
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data       interface{}            `json:"data,omitempty"`
	Errors     []GraphQLError         `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Locations  []GraphQLLocation      `json:"locations,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLLocation represents a GraphQL error location
type GraphQLLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// QueryExecutionPlan represents a plan for executing a federated query
type QueryExecutionPlan struct {
	QueryID       string              `json:"query_id"`
	ServiceCalls  []ServiceCall       `json:"service_calls"`
	Dependencies  map[string][]string `json:"dependencies"`
	EstimatedCost int                 `json:"estimated_cost"`
	Timeout       time.Duration       `json:"timeout"`
}

// ServiceCall represents a call to a specific service
type ServiceCall struct {
	ServiceName string                 `json:"service_name"`
	ServiceURL  string                 `json:"service_url"`
	Query       string                 `json:"query"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	RequiredBy  []string               `json:"required_by,omitempty"`
	Provides    []string               `json:"provides,omitempty"`
}

// CompositionOptions represents options for schema composition
type CompositionOptions struct {
	EnableQueryComplexityAnalysis bool     `json:"enable_query_complexity_analysis"`
	MaxQueryDepth                 int      `json:"max_query_depth"`
	MaxQueryComplexity            int      `json:"max_query_complexity"`
	EnableDistributedTracing      bool     `json:"enable_distributed_tracing"`
	EnableCaching                 bool     `json:"enable_caching"`
	AllowedServices               []string `json:"allowed_services,omitempty"`
}

// CompositionResult represents the result of schema composition
type CompositionResult struct {
	CompositionID    string                 `json:"composition_id"`
	ComposedSchema   string                 `json:"composed_schema"`
	Services         []string               `json:"services"`
	Version          string                 `json:"version"`
	CreatedAt        time.Time              `json:"created_at"`
	ValidationErrors []string               `json:"validation_errors,omitempty"`
	Warnings         []string               `json:"warnings,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ComposeSchema composes schemas from all healthy services
func (f *FederationGatewayService) ComposeSchema(ctx context.Context, options *CompositionOptions) (*CompositionResult, error) {
	f.logger.Info("Starting schema composition")

	// Get healthy services
	services, err := f.serviceDiscovery.GetHealthyServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy services: %w", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy services available for composition")
	}

	// Filter services if specified
	if len(options.AllowedServices) > 0 {
		services = f.filterAllowedServices(services, options.AllowedServices)
	}

	// Get schemas for each service
	var schemas []*domain.GraphQLSchema
	var serviceNames []string

	for _, service := range services {
		schema, err := f.schemaRepo.GetLatestByService(ctx, service.ServiceName)
		if err != nil {
			f.logger.Warn("Failed to get schema for service",
				zap.String("service", service.ServiceName),
				zap.Error(err),
			)
			continue
		}

		if schema != nil && schema.Status == domain.SchemaStatusActive {
			schemas = append(schemas, schema)
			serviceNames = append(serviceNames, service.ServiceName)
		}
	}

	if len(schemas) == 0 {
		return nil, fmt.Errorf("no valid schemas found for composition")
	}

	// Compose schemas
	composedSchema, validationErrors, warnings := f.performComposition(schemas, options)

	// Create composition record
	composition := &domain.FederationComposition{
		Services:         serviceNames,
		ComposedSchema:   composedSchema,
		Status:           domain.CompositionStatusActive,
		ValidationErrors: validationErrors,
		Warnings:         warnings,
		Configuration: map[string]interface{}{
			"max_query_depth":      options.MaxQueryDepth,
			"max_query_complexity": options.MaxQueryComplexity,
			"enable_caching":       options.EnableCaching,
			"enable_tracing":       options.EnableDistributedTracing,
		},
	}

	if len(validationErrors) > 0 {
		composition.Status = domain.CompositionStatusInvalid
	}

	if err := f.compositionRepo.Create(ctx, composition); err != nil {
		f.logger.Error("Failed to save composition", zap.Error(err))
		return nil, fmt.Errorf("failed to save composition: %w", err)
	}

	// Update cache
	f.updateCompositionCache(composition)

	// Create composition event
	changeEvent := &domain.SchemaChangeEvent{
		ServiceName: "gateway",
		ChangeType:  domain.ChangeTypeComposition,
		ChangeDetails: map[string]interface{}{
			"composition_id":    composition.ID.String(),
			"services":          serviceNames,
			"validation_errors": validationErrors,
			"warnings":          warnings,
		},
	}

	if err := f.changeEventRepo.Create(ctx, changeEvent); err != nil {
		f.logger.Error("Failed to create composition event", zap.Error(err))
	}

	f.logger.Info("Schema composition completed",
		zap.String("composition_id", composition.ID.String()),
		zap.Strings("services", serviceNames),
		zap.Int("validation_errors", len(validationErrors)),
		zap.Int("warnings", len(warnings)),
	)

	return &CompositionResult{
		CompositionID:    composition.ID.String(),
		ComposedSchema:   composedSchema,
		Services:         serviceNames,
		Version:          composition.Version,
		CreatedAt:        composition.CreatedAt,
		ValidationErrors: validationErrors,
		Warnings:         warnings,
		Metadata: map[string]interface{}{
			"service_count": len(serviceNames),
			"schema_size":   len(composedSchema),
		},
	}, nil
}

// ExecuteQuery executes a federated GraphQL query
func (f *FederationGatewayService) ExecuteQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	start := time.Now()
	queryID := f.generateQueryID()

	f.logger.Info("Executing federated query",
		zap.String("query_id", queryID),
		zap.String("operation", req.OperationName),
	)

	// Create execution plan
	plan, err := f.createExecutionPlan(ctx, req)
	if err != nil {
		return &GraphQLResponse{
			Errors: []GraphQLError{{
				Message: fmt.Sprintf("Failed to create execution plan: %v", err),
				Extensions: map[string]interface{}{
					"code": "EXECUTION_PLAN_ERROR",
				},
			}},
		}, nil
	}

	// Execute the plan
	response, err := f.executePlan(ctx, plan)
	if err != nil {
		return &GraphQLResponse{
			Errors: []GraphQLError{{
				Message: fmt.Sprintf("Query execution failed: %v", err),
				Extensions: map[string]interface{}{
					"code": "EXECUTION_ERROR",
				},
			}},
		}, nil
	}

	// Record metrics
	duration := time.Since(start)
	go f.recordQueryMetrics(context.Background(), queryID, req, plan, duration, response)

	f.logger.Info("Query execution completed",
		zap.String("query_id", queryID),
		zap.Duration("duration", duration),
		zap.Int("service_calls", len(plan.ServiceCalls)),
	)

	return response, nil
}

// GetActiveComposition returns the current active composition
func (f *FederationGatewayService) GetActiveComposition(ctx context.Context) (*domain.FederationComposition, error) {
	// Check cache first
	if cached, ok := f.compositionCache.Load("active"); ok {
		if composition, ok := cached.(*domain.FederationComposition); ok {
			return composition, nil
		}
	}

	composition, err := f.compositionRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	if composition != nil {
		f.compositionCache.Store("active", composition)
	}

	return composition, nil
}

// UpdateGatewayConfig updates the gateway configuration
func (f *FederationGatewayService) UpdateGatewayConfig(ctx context.Context, config map[string]interface{}) error {
	gatewayConfig := &domain.FederationGatewayConfig{
		ConfigKey:   "default",
		ConfigValue: config,
		IsActive:    true,
	}

	if err := f.gatewayConfigRepo.Upsert(ctx, gatewayConfig); err != nil {
		return fmt.Errorf("failed to update gateway config: %w", err)
	}

	// Update cache
	f.configCache.Store("default", gatewayConfig)

	return nil
}

// filterAllowedServices filters services based on allowed list
func (f *FederationGatewayService) filterAllowedServices(services []*domain.FederationService, allowed []string) []*domain.FederationService {
	allowedMap := make(map[string]bool)
	for _, name := range allowed {
		allowedMap[name] = true
	}

	var filtered []*domain.FederationService
	for _, service := range services {
		if allowedMap[service.ServiceName] {
			filtered = append(filtered, service)
		}
	}

	return filtered
}

// performComposition performs the actual schema composition
func (f *FederationGatewayService) performComposition(schemas []*domain.GraphQLSchema, options *CompositionOptions) (string, []string, []string) {
	var composedParts []string
	var validationErrors []string
	var warnings []string

	// Simple composition logic - in production, use proper federation
	for _, schema := range schemas {
		if schema.SchemaSDL != "" {
			composedParts = append(composedParts, schema.SchemaSDL)
		}
	}

	composedSchema := strings.Join(composedParts, "\n\n")

	// Basic validation
	if len(composedSchema) == 0 {
		validationErrors = append(validationErrors, "Empty composed schema")
	}

	if len(schemas) > 10 {
		warnings = append(warnings, "Large number of services may impact performance")
	}

	return composedSchema, validationErrors, warnings
}

// createExecutionPlan creates an execution plan for a query
func (f *FederationGatewayService) createExecutionPlan(ctx context.Context, req *GraphQLRequest) (*QueryExecutionPlan, error) {
	services, err := f.serviceDiscovery.GetHealthyServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	var serviceCalls []ServiceCall
	for _, service := range services {
		serviceCalls = append(serviceCalls, ServiceCall{
			ServiceName: service.ServiceName,
			ServiceURL:  service.ServiceURL,
			Query:       req.Query,
			Variables:   req.Variables,
		})
	}

	return &QueryExecutionPlan{
		QueryID:       f.generateQueryID(),
		ServiceCalls:  serviceCalls,
		Dependencies:  make(map[string][]string),
		EstimatedCost: len(serviceCalls) * 10,
		Timeout:       30 * time.Second,
	}, nil
}

// executePlan executes a query plan
func (f *FederationGatewayService) executePlan(ctx context.Context, plan *QueryExecutionPlan) (*GraphQLResponse, error) {
	// Simple execution - in production, implement proper federation
	response := &GraphQLResponse{
		Data: map[string]interface{}{
			"federation": map[string]interface{}{
				"services": len(plan.ServiceCalls),
				"queryId":  plan.QueryID,
			},
		},
	}

	return response, nil
}

// recordQueryMetrics records query execution metrics
func (f *FederationGatewayService) recordQueryMetrics(ctx context.Context, queryID string, req *GraphQLRequest, plan *QueryExecutionPlan, duration time.Duration, response *GraphQLResponse) {
	metrics := &domain.GraphQLQueryMetrics{
		QueryHash:       f.hashQuery(req.Query),
		ServicesCalled:  f.extractServiceNames(plan.ServiceCalls),
		ExecutionTime:   duration,
		QueryComplexity: plan.EstimatedCost,
		CacheHit:        false,
		ErrorCount:      len(response.Errors),
		Metadata: map[string]interface{}{
			"query_id":       queryID,
			"operation_name": req.OperationName,
			"service_count":  len(plan.ServiceCalls),
		},
	}

	if err := f.queryMetricsRepo.Create(ctx, metrics); err != nil {
		f.logger.Error("Failed to record query metrics",
			zap.Error(err),
			zap.String("query_id", queryID),
		)
	}
}

// updateCompositionCache updates the composition cache
func (f *FederationGatewayService) updateCompositionCache(composition *domain.FederationComposition) {
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()

	f.compositionCache.Store("active", composition)
	f.lastCacheUpdate = time.Now()
}

// generateQueryID generates a unique query ID
func (f *FederationGatewayService) generateQueryID() string {
	return fmt.Sprintf("query_%d", time.Now().UnixNano())
}

// hashQuery creates a hash of the query for metrics
func (f *FederationGatewayService) hashQuery(query string) string {
	// Simple hash - in production, use proper hashing
	return fmt.Sprintf("hash_%d", len(query))
}

// extractServiceNames extracts service names from service calls
func (f *FederationGatewayService) extractServiceNames(calls []ServiceCall) []string {
	var names []string
	for _, call := range calls {
		names = append(names, call.ServiceName)
	}
	return names
}
