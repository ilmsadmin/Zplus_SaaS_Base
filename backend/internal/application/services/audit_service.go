package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// AuditServiceImpl implements AuditService
type AuditServiceImpl struct {
	auditRepo domain.AuditLogRepository
}

// NewAuditService creates a new audit service
func NewAuditService(auditRepo domain.AuditLogRepository) AuditService {
	return &AuditServiceImpl{
		auditRepo: auditRepo,
	}
}

// LogEvent logs an audit event
func (s *AuditServiceImpl) LogEvent(
	ctx context.Context,
	tenantID uuid.UUID,
	userID *uuid.UUID,
	action, resource, resourceID string,
	details map[string]interface{},
) error {
	auditLog := &domain.AuditLog{
		ID:         uuid.New(),
		TenantID:   tenantID.String(),
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: &resourceID,
		Details: &domain.Details{
			Metadata: details,
		},
		CreatedAt: time.Now(),
	}

	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}
