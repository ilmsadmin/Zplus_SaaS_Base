package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// SSLCertificateRequestRepositoryImpl implements SSLCertificateRequestRepository
type SSLCertificateRequestRepositoryImpl struct {
	db *gorm.DB
}

func NewSSLCertificateRequestRepository(db *gorm.DB) *SSLCertificateRequestRepositoryImpl {
	return &SSLCertificateRequestRepositoryImpl{db: db}
}

func (r *SSLCertificateRequestRepositoryImpl) Create(ctx context.Context, request *domain.SSLCertificateRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *SSLCertificateRequestRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.SSLCertificateRequest, error) {
	var request domain.SSLCertificateRequest
	err := r.db.WithContext(ctx).First(&request, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *SSLCertificateRequestRepositoryImpl) GetByDomainID(ctx context.Context, domainID uuid.UUID) ([]*domain.SSLCertificateRequest, error) {
	var requests []*domain.SSLCertificateRequest
	err := r.db.WithContext(ctx).Where("domain_id = ?", domainID).Order("created_at DESC").Find(&requests).Error
	return requests, err
}

func (r *SSLCertificateRequestRepositoryImpl) GetByStatus(ctx context.Context, status string) ([]*domain.SSLCertificateRequest, error) {
	var requests []*domain.SSLCertificateRequest
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&requests).Error
	return requests, err
}

func (r *SSLCertificateRequestRepositoryImpl) Update(ctx context.Context, request *domain.SSLCertificateRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}

func (r *SSLCertificateRequestRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.SSLCertificateRequest{}, "id = ?", id).Error
}

func (r *SSLCertificateRequestRepositoryImpl) GetPendingRequests(ctx context.Context) ([]*domain.SSLCertificateRequest, error) {
	var requests []*domain.SSLCertificateRequest
	err := r.db.WithContext(ctx).Where("status IN ? AND attempt_count < max_attempts",
		[]string{"pending", "processing"}).Find(&requests).Error
	return requests, err
}

func (r *SSLCertificateRequestRepositoryImpl) GetExpiredRequests(ctx context.Context) ([]*domain.SSLCertificateRequest, error) {
	var requests []*domain.SSLCertificateRequest
	err := r.db.WithContext(ctx).Where("expires_at <= ? AND status NOT IN ?",
		time.Now(), []string{"completed", "failed"}).Find(&requests).Error
	return requests, err
}

func (r *SSLCertificateRequestRepositoryImpl) IncrementAttemptCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.SSLCertificateRequest{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"attempt_count":   gorm.Expr("attempt_count + 1"),
			"last_attempt_at": time.Now(),
		}).Error
}

func (r *SSLCertificateRequestRepositoryImpl) MarkAsCompleted(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.SSLCertificateRequest{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": time.Now(),
		}).Error
}

func (r *SSLCertificateRequestRepositoryImpl) UpdateChallengeData(ctx context.Context, id uuid.UUID, challengeData map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&domain.SSLCertificateRequest{}).Where("id = ?", id).
		Update("challenge_data", challengeData).Error
}
