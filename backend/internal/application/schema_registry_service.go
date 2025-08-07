package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// SchemaRegistryService handles GraphQL schema registration and management
type SchemaRegistryService struct {
	schemaRepo      domain.GraphQLSchemaRepository
	serviceRepo     domain.FederationServiceRepository
	compositionRepo domain.FederationCompositionRepository
	changeEventRepo domain.SchemaChangeEventRepository
	logger          *zap.Logger
}

// NewSchemaRegistryService creates a new schema registry service
func NewSchemaRegistryService(
	schemaRepo domain.GraphQLSchemaRepository,
	serviceRepo domain.FederationServiceRepository,
	compositionRepo domain.FederationCompositionRepository,
	changeEventRepo domain.SchemaChangeEventRepository,
	logger *zap.Logger,
) *SchemaRegistryService {
	return &SchemaRegistryService{
		schemaRepo:      schemaRepo,
		serviceRepo:     serviceRepo,
		compositionRepo: compositionRepo,
		changeEventRepo: changeEventRepo,
		logger:          logger,
	}
}

// RegisterSchemaRequest represents a schema registration request
type RegisterSchemaRequest struct {
	ServiceName    string                 `json:"service_name" validate:"required"`
	ServiceVersion string                 `json:"service_version" validate:"required"`
	SchemaSDL      string                 `json:"schema_sdl" validate:"required"`
	ServiceURL     string                 `json:"service_url" validate:"required"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// RegisterSchemaResponse represents a schema registration response
type RegisterSchemaResponse struct {
	SchemaID         string   `json:"schema_id"`
	SchemaHash       string   `json:"schema_hash"`
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
	BreakingChanges  []string `json:"breaking_changes,omitempty"`
	ServiceURL       string   `json:"service_url"`
	RegisteredAt     string   `json:"registered_at"`
}

// RegisterSchema registers a new GraphQL schema for a service
func (s *SchemaRegistryService) RegisterSchema(ctx context.Context, req *RegisterSchemaRequest) (*RegisterSchemaResponse, error) {
	s.logger.Info("Registering GraphQL schema",
		zap.String("service", req.ServiceName),
		zap.String("version", req.ServiceVersion),
	)

	// Calculate schema hash
	schemaHash := s.calculateSchemaHash(req.SchemaSDL)

	// Check if schema already exists
	existingSchema, err := s.schemaRepo.GetByServiceAndVersion(ctx, req.ServiceName, req.ServiceVersion)
	if err == nil && existingSchema != nil {
		if existingSchema.SchemaHash == schemaHash {
			s.logger.Info("Schema already registered with same hash",
				zap.String("service", req.ServiceName),
				zap.String("version", req.ServiceVersion),
				zap.String("hash", schemaHash),
			)
			return &RegisterSchemaResponse{
				SchemaID:     existingSchema.ID.String(),
				SchemaHash:   existingSchema.SchemaHash,
				IsValid:      existingSchema.IsValid,
				ServiceURL:   req.ServiceURL,
				RegisteredAt: existingSchema.CreatedAt.Format(time.RFC3339),
			}, nil
		}
	}

	// Validate schema SDL
	validationErrors := s.validateSchemaSDL(req.SchemaSDL)
	isValid := len(validationErrors) == 0

	// Create schema record
	schema := &domain.GraphQLSchema{
		ServiceName:      req.ServiceName,
		ServiceVersion:   req.ServiceVersion,
		SchemaSDL:        req.SchemaSDL,
		SchemaHash:       schemaHash,
		IsActive:         isValid,
		IsValid:          isValid,
		ValidationErrors: validationErrors,
		Metadata:         req.Metadata,
	}

	if err := s.schemaRepo.Create(ctx, schema); err != nil {
		s.logger.Error("Failed to create schema record",
			zap.Error(err),
			zap.String("service", req.ServiceName),
		)
		return nil, fmt.Errorf("failed to create schema record: %w", err)
	}

	// Register or update service
	service, err := s.serviceRepo.GetByName(ctx, req.ServiceName)
	if err != nil {
		// Create new service
		service = &domain.FederationService{
			ServiceName:    req.ServiceName,
			ServiceURL:     req.ServiceURL,
			HealthCheckURL: req.ServiceURL + "/health",
			SchemaID:       &schema.ID,
			Status:         domain.ServiceStatusUnknown,
			Weight:         100,
			Tags:           []string{"graphql", "federation"},
			Metadata: map[string]interface{}{
				"version": req.ServiceVersion,
			},
		}

		if err := s.serviceRepo.Register(ctx, service); err != nil {
			s.logger.Error("Failed to register service",
				zap.Error(err),
				zap.String("service", req.ServiceName),
			)
		}
	} else {
		// Update existing service
		service.ServiceURL = req.ServiceURL
		service.SchemaID = &schema.ID
		service.Metadata["version"] = req.ServiceVersion

		if err := s.serviceRepo.Update(ctx, service); err != nil {
			s.logger.Error("Failed to update service",
				zap.Error(err),
				zap.String("service", req.ServiceName),
			)
		}
	}

	// Check for breaking changes
	var breakingChanges []string
	if existingSchema != nil {
		breakingChanges = s.detectBreakingChanges(existingSchema.SchemaSDL, req.SchemaSDL)
	}

	// Create change event
	changeEvent := &domain.SchemaChangeEvent{
		ServiceName:     req.ServiceName,
		ChangeType:      domain.ChangeTypeSchemaUpdated,
		OldSchemaID:     getSchemaIDPtr(existingSchema),
		NewSchemaID:     &schema.ID,
		BreakingChanges: breakingChanges,
		ChangeDetails: map[string]interface{}{
			"new_version":       req.ServiceVersion,
			"validation_errors": validationErrors,
			"is_valid":          isValid,
			"breaking_changes":  breakingChanges,
		},
	}

	if err := s.changeEventRepo.Create(ctx, changeEvent); err != nil {
		s.logger.Error("Failed to create change event",
			zap.Error(err),
			zap.String("service", req.ServiceName),
		)
	}

	s.logger.Info("Schema registration completed",
		zap.String("service", req.ServiceName),
		zap.String("version", req.ServiceVersion),
		zap.String("schema_id", schema.ID.String()),
		zap.Bool("is_valid", isValid),
		zap.Int("breaking_changes", len(breakingChanges)),
	)

	return &RegisterSchemaResponse{
		SchemaID:         schema.ID.String(),
		SchemaHash:       schemaHash,
		IsValid:          isValid,
		ValidationErrors: validationErrors,
		BreakingChanges:  breakingChanges,
		ServiceURL:       req.ServiceURL,
		RegisteredAt:     schema.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetSchema retrieves a schema by service name and version
func (s *SchemaRegistryService) GetSchema(ctx context.Context, serviceName, version string) (*domain.GraphQLSchema, error) {
	if version == "latest" {
		return s.schemaRepo.GetLatestByService(ctx, serviceName)
	}
	return s.schemaRepo.GetByServiceAndVersion(ctx, serviceName, version)
}

// GetActiveSchemas returns all currently active schemas
func (s *SchemaRegistryService) GetActiveSchemas(ctx context.Context) ([]*domain.GraphQLSchema, error) {
	return s.schemaRepo.GetActiveSchemas(ctx)
}

// DeregisterService removes a service from the registry
func (s *SchemaRegistryService) DeregisterService(ctx context.Context, serviceName string) error {
	s.logger.Info("Deregistering service", zap.String("service", serviceName))

	service, err := s.serviceRepo.GetByName(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// Create deregistration event
	changeEvent := &domain.SchemaChangeEvent{
		ServiceName: serviceName,
		ChangeType:  domain.ChangeTypeServiceDeregistered,
		OldSchemaID: service.SchemaID,
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

// calculateSchemaHash calculates SHA256 hash of schema SDL
func (s *SchemaRegistryService) calculateSchemaHash(schemaSDL string) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(schemaSDL)))
	return hex.EncodeToString(hash[:])
}

// validateSchemaSDL validates GraphQL schema SDL
func (s *SchemaRegistryService) validateSchemaSDL(schemaSDL string) []string {
	var errors []string

	// Parse the schema
	doc, parseErr := parser.ParseSchema(&ast.Source{
		Input: schemaSDL,
	})

	// Basic validation - just parse the schema
	if parseErr != nil {
		errors = append(errors, fmt.Sprintf("Parse error: %s", parseErr.Error()))
		return errors
	}

	// Schema parsed successfully - basic validation passed
	s.logger.Debug("Schema validation completed",
		zap.Int("definition_count", len(doc.Definitions)),
	)

	// Validate federation directives
	federationErrors := s.validateFederationDirectives(doc)
	errors = append(errors, federationErrors...)

	return errors
}

// validateFederationDirectives validates Apollo Federation directives
func (s *SchemaRegistryService) validateFederationDirectives(doc *ast.SchemaDocument) []string {
	var errors []string

	// Check for required federation directives
	hasKey := false
	hasEntity := false

	for _, def := range doc.Definitions {
		if def.Kind == ast.Object {
			for _, directive := range def.Directives {
				switch directive.Name {
				case "key":
					hasKey = true
				case "entity":
					hasEntity = true
				}
			}
		}
	}

	// Additional federation-specific validations can be added here
	_ = hasKey
	_ = hasEntity

	return errors
}

// detectBreakingChanges detects breaking changes between schema versions
func (s *SchemaRegistryService) detectBreakingChanges(oldSDL, newSDL string) []string {
	var breakingChanges []string

	// Parse both schemas
	oldDoc, err := parser.ParseSchema(&ast.Source{Input: oldSDL})
	if err != nil {
		return []string{"Failed to parse old schema"}
	}

	newDoc, err := parser.ParseSchema(&ast.Source{Input: newSDL})
	if err != nil {
		return []string{"Failed to parse new schema"}
	}

	// Compare schemas for breaking changes
	oldTypes := make(map[string]*ast.Definition)
	newTypes := make(map[string]*ast.Definition)

	for _, def := range oldDoc.Definitions {
		oldTypes[def.Name] = def
	}

	for _, def := range newDoc.Definitions {
		newTypes[def.Name] = def
	}

	// Check for removed types
	for typeName := range oldTypes {
		if _, exists := newTypes[typeName]; !exists {
			breakingChanges = append(breakingChanges, fmt.Sprintf("Type '%s' was removed", typeName))
		}
	}

	// Check for field changes in existing types
	for typeName, oldType := range oldTypes {
		if newType, exists := newTypes[typeName]; exists {
			fieldChanges := s.compareTypeFields(oldType, newType)
			breakingChanges = append(breakingChanges, fieldChanges...)
		}
	}

	return breakingChanges
}

// compareTypeFields compares fields between two type definitions
func (s *SchemaRegistryService) compareTypeFields(oldType, newType *ast.Definition) []string {
	var changes []string

	if oldType.Kind != newType.Kind {
		return []string{fmt.Sprintf("Type '%s' kind changed from %s to %s", oldType.Name, oldType.Kind, newType.Kind)}
	}

	if oldType.Kind == ast.Object || oldType.Kind == ast.Interface {
		oldFields := make(map[string]*ast.FieldDefinition)
		newFields := make(map[string]*ast.FieldDefinition)

		for _, field := range oldType.Fields {
			oldFields[field.Name] = field
		}

		for _, field := range newType.Fields {
			newFields[field.Name] = field
		}

		// Check for removed fields
		for fieldName := range oldFields {
			if _, exists := newFields[fieldName]; !exists {
				changes = append(changes, fmt.Sprintf("Field '%s.%s' was removed", oldType.Name, fieldName))
			}
		}

		// Check for field type changes
		for fieldName, oldField := range oldFields {
			if newField, exists := newFields[fieldName]; exists {
				if oldField.Type.String() != newField.Type.String() {
					changes = append(changes, fmt.Sprintf("Field '%s.%s' type changed from %s to %s",
						oldType.Name, fieldName, oldField.Type.String(), newField.Type.String()))
				}
			}
		}
	}

	return changes
}

// Helper function to get schema ID pointer
func getSchemaIDPtr(schema *domain.GraphQLSchema) *uuid.UUID {
	if schema == nil {
		return nil
	}
	return &schema.ID
}
