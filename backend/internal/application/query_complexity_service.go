package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// QueryComplexityService analyzes and validates GraphQL query complexity
type QueryComplexityService struct {
	schemaRepo       domain.GraphQLSchemaRepository
	queryMetricsRepo domain.GraphQLQueryMetricsRepository
	logger           *zap.Logger

	// Configuration
	maxDepth      int
	maxComplexity int
	enableCaching bool
}

// NewQueryComplexityService creates a new query complexity service
func NewQueryComplexityService(
	schemaRepo domain.GraphQLSchemaRepository,
	queryMetricsRepo domain.GraphQLQueryMetricsRepository,
	logger *zap.Logger,
) *QueryComplexityService {
	return &QueryComplexityService{
		schemaRepo:       schemaRepo,
		queryMetricsRepo: queryMetricsRepo,
		logger:           logger,
		maxDepth:         15,   // Default max depth
		maxComplexity:    1000, // Default max complexity
		enableCaching:    true,
	}
}

// ComplexityAnalysis represents the result of query complexity analysis
type ComplexityAnalysis struct {
	QueryHash        string                 `json:"query_hash"`
	Depth            int                    `json:"depth"`
	Complexity       int                    `json:"complexity"`
	FieldCount       int                    `json:"field_count"`
	FragmentCount    int                    `json:"fragment_count"`
	OperationType    string                 `json:"operation_type"`
	IsValid          bool                   `json:"is_valid"`
	Violations       []ComplexityViolation  `json:"violations,omitempty"`
	Suggestions      []string               `json:"suggestions,omitempty"`
	EstimatedCost    int                    `json:"estimated_cost"`
	CacheRecommended bool                   `json:"cache_recommended"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ComplexityViolation represents a complexity rule violation
type ComplexityViolation struct {
	Type        string `json:"type"`
	Field       string `json:"field,omitempty"`
	Path        string `json:"path,omitempty"`
	Current     int    `json:"current"`
	Limit       int    `json:"limit"`
	Description string `json:"description"`
}

// ComplexityConfig represents complexity analysis configuration
type ComplexityConfig struct {
	MaxDepth            int                    `json:"max_depth"`
	MaxComplexity       int                    `json:"max_complexity"`
	MaxFieldCount       int                    `json:"max_field_count"`
	FieldCosts          map[string]int         `json:"field_costs"`
	TypeCosts           map[string]int         `json:"type_costs"`
	EnableIntrospection bool                   `json:"enable_introspection"`
	CustomRules         []CustomComplexityRule `json:"custom_rules"`
}

// CustomComplexityRule represents a custom complexity rule
type CustomComplexityRule struct {
	Name        string                 `json:"name"`
	Pattern     string                 `json:"pattern"`
	CostFactor  int                    `json:"cost_factor"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// QueryComplexityRequest represents a request to analyze query complexity
type QueryComplexityRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operation_name,omitempty"`
	SchemaSDL     string                 `json:"schema_sdl,omitempty"`
	ServiceName   string                 `json:"service_name,omitempty"`
	Config        *ComplexityConfig      `json:"config,omitempty"`
}

// AnalyzeComplexity analyzes the complexity of a GraphQL query
func (q *QueryComplexityService) AnalyzeComplexity(ctx context.Context, req *QueryComplexityRequest) (*ComplexityAnalysis, error) {
	q.logger.Info("Analyzing query complexity",
		zap.String("service", req.ServiceName),
		zap.String("operation", req.OperationName),
	)

	// Parse the query
	queryDoc, err := parser.ParseQuery(&ast.Source{Input: req.Query})
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Get or use provided schema
	var schemaDoc *ast.SchemaDocument
	if req.SchemaSDL != "" {
		schemaDoc, err = parser.ParseSchema(&ast.Source{Input: req.SchemaSDL})
		if err != nil {
			return nil, fmt.Errorf("failed to parse schema: %w", err)
		}
	} else if req.ServiceName != "" {
		schema, err := q.schemaRepo.GetLatestByService(ctx, req.ServiceName)
		if err != nil {
			return nil, fmt.Errorf("failed to get schema for service %s: %w", req.ServiceName, err)
		}
		schemaDoc, err = parser.ParseSchema(&ast.Source{Input: schema.SchemaSDL})
		if err != nil {
			return nil, fmt.Errorf("failed to parse service schema: %w", err)
		}
	}

	// Validate query against schema if available
	if schemaDoc != nil {
		// Basic validation - just check if schema was parsed successfully
		q.logger.Debug("Schema available for validation",
			zap.Int("definition_count", len(schemaDoc.Definitions)),
		)
	}

	// Apply configuration
	config := q.getEffectiveConfig(req.Config)

	// Perform complexity analysis
	analysis := &ComplexityAnalysis{
		QueryHash:     q.generateQueryHash(req.Query),
		OperationType: q.getOperationType(queryDoc),
		IsValid:       true,
		Violations:    []ComplexityViolation{},
		Suggestions:   []string{},
		Metadata:      make(map[string]interface{}),
	}

	// Analyze depth
	analysis.Depth = q.calculateDepth(queryDoc)
	if analysis.Depth > config.MaxDepth {
		analysis.IsValid = false
		analysis.Violations = append(analysis.Violations, ComplexityViolation{
			Type:        "max_depth_exceeded",
			Current:     analysis.Depth,
			Limit:       config.MaxDepth,
			Description: fmt.Sprintf("Query depth %d exceeds maximum allowed depth %d", analysis.Depth, config.MaxDepth),
		})
		analysis.Suggestions = append(analysis.Suggestions, "Consider reducing query nesting levels")
	}

	// Analyze field count
	analysis.FieldCount = q.calculateFieldCount(queryDoc)
	if config.MaxFieldCount > 0 && analysis.FieldCount > config.MaxFieldCount {
		analysis.IsValid = false
		analysis.Violations = append(analysis.Violations, ComplexityViolation{
			Type:        "max_field_count_exceeded",
			Current:     analysis.FieldCount,
			Limit:       config.MaxFieldCount,
			Description: fmt.Sprintf("Query field count %d exceeds maximum allowed %d", analysis.FieldCount, config.MaxFieldCount),
		})
		analysis.Suggestions = append(analysis.Suggestions, "Consider splitting query into multiple smaller queries")
	}

	// Calculate complexity score
	analysis.Complexity = q.calculateComplexity(queryDoc, schemaDoc, config)
	if analysis.Complexity > config.MaxComplexity {
		analysis.IsValid = false
		analysis.Violations = append(analysis.Violations, ComplexityViolation{
			Type:        "max_complexity_exceeded",
			Current:     analysis.Complexity,
			Limit:       config.MaxComplexity,
			Description: fmt.Sprintf("Query complexity %d exceeds maximum allowed %d", analysis.Complexity, config.MaxComplexity),
		})
		analysis.Suggestions = append(analysis.Suggestions, "Consider reducing the number of fields or using pagination")
	}

	// Count fragments
	analysis.FragmentCount = q.calculateFragmentCount(queryDoc)

	// Calculate estimated cost
	analysis.EstimatedCost = q.calculateEstimatedCost(analysis)

	// Check if caching is recommended
	analysis.CacheRecommended = q.shouldRecommendCaching(analysis)

	// Add metadata
	analysis.Metadata["analysis_time"] = time.Now().Format(time.RFC3339)
	analysis.Metadata["config_applied"] = config
	analysis.Metadata["schema_available"] = schemaDoc != nil

	q.logger.Info("Query complexity analysis completed",
		zap.String("query_hash", analysis.QueryHash),
		zap.Int("depth", analysis.Depth),
		zap.Int("complexity", analysis.Complexity),
		zap.Int("field_count", analysis.FieldCount),
		zap.Bool("is_valid", analysis.IsValid),
		zap.Int("violations", len(analysis.Violations)),
	)

	return analysis, nil
}

// UpdateConfig updates the complexity analysis configuration
func (q *QueryComplexityService) UpdateConfig(config *ComplexityConfig) {
	if config.MaxDepth > 0 {
		q.maxDepth = config.MaxDepth
	}
	if config.MaxComplexity > 0 {
		q.maxComplexity = config.MaxComplexity
	}

	q.logger.Info("Query complexity config updated",
		zap.Int("max_depth", q.maxDepth),
		zap.Int("max_complexity", q.maxComplexity),
	)
}

// GetQueryStats returns query complexity statistics
func (q *QueryComplexityService) GetQueryStats(ctx context.Context, queryHash string, hours int) (*ComplexityAnalysis, error) {
	// Get recent metrics for the query
	metrics, err := q.queryMetricsRepo.GetByQueryHash(ctx, queryHash, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get query metrics: %w", err)
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no metrics found for query hash %s", queryHash)
	}

	// Calculate aggregate stats
	return q.calculateAggregateStats(metrics), nil
}

// calculateDepth calculates the maximum depth of a query
func (q *QueryComplexityService) calculateDepth(queryDoc *ast.QueryDocument) int {
	maxDepth := 0

	for _, operation := range queryDoc.Operations {
		depth := q.calculateSelectionDepth(operation.SelectionSet, 1)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// calculateSelectionDepth calculates the depth of a selection set
func (q *QueryComplexityService) calculateSelectionDepth(selections ast.SelectionSet, currentDepth int) int {
	maxDepth := currentDepth

	for _, selection := range selections {
		switch sel := selection.(type) {
		case *ast.Field:
			if len(sel.SelectionSet) > 0 {
				depth := q.calculateSelectionDepth(sel.SelectionSet, currentDepth+1)
				if depth > maxDepth {
					maxDepth = depth
				}
			}
		case *ast.InlineFragment:
			depth := q.calculateSelectionDepth(sel.SelectionSet, currentDepth)
			if depth > maxDepth {
				maxDepth = depth
			}
		case *ast.FragmentSpread:
			// Fragment depth would need fragment definition resolution
			// For now, assume +1 depth
			depth := currentDepth + 1
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

// calculateFieldCount calculates the total number of fields in a query
func (q *QueryComplexityService) calculateFieldCount(queryDoc *ast.QueryDocument) int {
	totalFields := 0

	for _, operation := range queryDoc.Operations {
		totalFields += q.countSelectionFields(operation.SelectionSet)
	}

	return totalFields
}

// countSelectionFields counts fields in a selection set
func (q *QueryComplexityService) countSelectionFields(selections ast.SelectionSet) int {
	count := 0

	for _, selection := range selections {
		switch sel := selection.(type) {
		case *ast.Field:
			count++
			if len(sel.SelectionSet) > 0 {
				count += q.countSelectionFields(sel.SelectionSet)
			}
		case *ast.InlineFragment:
			count += q.countSelectionFields(sel.SelectionSet)
		case *ast.FragmentSpread:
			count++ // Count as one field for simplicity
		}
	}

	return count
}

// calculateComplexity calculates the complexity score of a query
func (q *QueryComplexityService) calculateComplexity(queryDoc *ast.QueryDocument, schemaDoc *ast.SchemaDocument, config *ComplexityConfig) int {
	totalComplexity := 0

	for _, operation := range queryDoc.Operations {
		totalComplexity += q.calculateSelectionComplexity(operation.SelectionSet, config, 1)
	}

	return totalComplexity
}

// calculateSelectionComplexity calculates complexity for a selection set
func (q *QueryComplexityService) calculateSelectionComplexity(selections ast.SelectionSet, config *ComplexityConfig, multiplier int) int {
	complexity := 0

	for _, selection := range selections {
		switch sel := selection.(type) {
		case *ast.Field:
			fieldCost := q.getFieldCost(sel.Name, config)
			fieldComplexity := fieldCost * multiplier

			if len(sel.SelectionSet) > 0 {
				// Nested fields multiply complexity
				nestedMultiplier := multiplier
				if q.isListField(sel) {
					nestedMultiplier *= 10 // List fields increase complexity
				}
				fieldComplexity += q.calculateSelectionComplexity(sel.SelectionSet, config, nestedMultiplier)
			}

			complexity += fieldComplexity
		case *ast.InlineFragment:
			complexity += q.calculateSelectionComplexity(sel.SelectionSet, config, multiplier)
		case *ast.FragmentSpread:
			complexity += 1 * multiplier // Base cost for fragment
		}
	}

	return complexity
}

// calculateFragmentCount counts the number of fragments in a query
func (q *QueryComplexityService) calculateFragmentCount(queryDoc *ast.QueryDocument) int {
	return len(queryDoc.Fragments)
}

// calculateEstimatedCost estimates the computational cost of a query
func (q *QueryComplexityService) calculateEstimatedCost(analysis *ComplexityAnalysis) int {
	// Base cost calculation
	cost := analysis.Complexity

	// Add depth penalty
	if analysis.Depth > 5 {
		cost += (analysis.Depth - 5) * 10
	}

	// Add field count penalty
	if analysis.FieldCount > 20 {
		cost += (analysis.FieldCount - 20) * 2
	}

	return cost
}

// shouldRecommendCaching determines if caching should be recommended
func (q *QueryComplexityService) shouldRecommendCaching(analysis *ComplexityAnalysis) bool {
	// Recommend caching for complex queries or queries with many fields
	return analysis.Complexity > 100 || analysis.FieldCount > 10 || analysis.Depth > 5
}

// getEffectiveConfig merges provided config with defaults
func (q *QueryComplexityService) getEffectiveConfig(config *ComplexityConfig) *ComplexityConfig {
	effective := &ComplexityConfig{
		MaxDepth:            q.maxDepth,
		MaxComplexity:       q.maxComplexity,
		MaxFieldCount:       50, // Default max field count
		FieldCosts:          make(map[string]int),
		TypeCosts:           make(map[string]int),
		EnableIntrospection: false,
		CustomRules:         []CustomComplexityRule{},
	}

	if config != nil {
		if config.MaxDepth > 0 {
			effective.MaxDepth = config.MaxDepth
		}
		if config.MaxComplexity > 0 {
			effective.MaxComplexity = config.MaxComplexity
		}
		if config.MaxFieldCount > 0 {
			effective.MaxFieldCount = config.MaxFieldCount
		}
		if config.FieldCosts != nil {
			effective.FieldCosts = config.FieldCosts
		}
		if config.TypeCosts != nil {
			effective.TypeCosts = config.TypeCosts
		}
		effective.EnableIntrospection = config.EnableIntrospection
		if config.CustomRules != nil {
			effective.CustomRules = config.CustomRules
		}
	}

	return effective
}

// getFieldCost returns the cost of a specific field
func (q *QueryComplexityService) getFieldCost(fieldName string, config *ComplexityConfig) int {
	if cost, exists := config.FieldCosts[fieldName]; exists {
		return cost
	}

	// Default field costs based on common patterns
	switch {
	case strings.HasSuffix(fieldName, "Connection"):
		return 10 // Connections are more expensive
	case strings.HasPrefix(fieldName, "search"):
		return 15 // Search operations are expensive
	case fieldName == "id" || fieldName == "createdAt" || fieldName == "updatedAt":
		return 1 // Simple scalar fields
	default:
		return 2 // Default field cost
	}
}

// isListField determines if a field returns a list
func (q *QueryComplexityService) isListField(field *ast.Field) bool {
	// This is a simplified check - in a real implementation,
	// you would use the schema to determine the field type
	return strings.Contains(strings.ToLower(field.Name), "list") ||
		strings.HasSuffix(field.Name, "s") ||
		strings.Contains(strings.ToLower(field.Name), "connection")
}

// getOperationType determines the operation type from the query document
func (q *QueryComplexityService) getOperationType(queryDoc *ast.QueryDocument) string {
	if len(queryDoc.Operations) == 0 {
		return "unknown"
	}

	operation := queryDoc.Operations[0]
	return string(operation.Operation)
}

// generateQueryHash generates a hash for the query
func (q *QueryComplexityService) generateQueryHash(query string) string {
	// Simple hash based on query length and first/last characters
	// In production, use a proper hash function
	normalized := strings.TrimSpace(strings.ReplaceAll(query, "\n", " "))
	return fmt.Sprintf("hash_%d_%s_%s",
		len(normalized),
		string(normalized[0]),
		string(normalized[len(normalized)-1]))
}

// calculateAggregateStats calculates aggregate statistics from metrics
func (q *QueryComplexityService) calculateAggregateStats(metrics []*domain.GraphQLQueryMetrics) *ComplexityAnalysis {
	if len(metrics) == 0 {
		return nil
	}

	// Calculate averages and aggregates
	totalComplexity := 0
	totalExecutionTime := int64(0)
	errorCount := 0

	for _, metric := range metrics {
		totalComplexity += metric.QueryComplexity
		totalExecutionTime += int64(metric.ExecutionTime)
		errorCount += metric.ErrorCount
	}

	return &ComplexityAnalysis{
		QueryHash:     metrics[0].QueryHash,
		Complexity:    totalComplexity / len(metrics),
		EstimatedCost: totalComplexity / len(metrics),
		Metadata: map[string]interface{}{
			"sample_count":       len(metrics),
			"avg_execution_time": time.Duration(totalExecutionTime / int64(len(metrics))),
			"total_errors":       errorCount,
			"error_rate":         float64(errorCount) / float64(len(metrics)),
		},
	}
}
