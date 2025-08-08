package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// DNSRecordRepositoryImpl implements DNSRecordRepository
type DNSRecordRepositoryImpl struct {
	db *gorm.DB
}

func NewDNSRecordRepository(db *gorm.DB) *DNSRecordRepositoryImpl {
	return &DNSRecordRepositoryImpl{db: db}
}

func (r *DNSRecordRepositoryImpl) Create(ctx context.Context, record *domain.DNSRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *DNSRecordRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.DNSRecord, error) {
	var record domain.DNSRecord
	err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *DNSRecordRepositoryImpl) GetByDomainID(ctx context.Context, domainID uuid.UUID) ([]*domain.DNSRecord, error) {
	var records []*domain.DNSRecord
	err := r.db.WithContext(ctx).Where("domain_id = ?", domainID).Find(&records).Error
	return records, err
}

func (r *DNSRecordRepositoryImpl) GetByType(ctx context.Context, domainID uuid.UUID, recordType string) ([]*domain.DNSRecord, error) {
	var records []*domain.DNSRecord
	err := r.db.WithContext(ctx).Where("domain_id = ? AND type = ?", domainID, recordType).Find(&records).Error
	return records, err
}

func (r *DNSRecordRepositoryImpl) GetByNameAndType(ctx context.Context, domainID uuid.UUID, name, recordType string) (*domain.DNSRecord, error) {
	var record domain.DNSRecord
	err := r.db.WithContext(ctx).Where("domain_id = ? AND name = ? AND type = ?", domainID, name, recordType).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *DNSRecordRepositoryImpl) Update(ctx context.Context, record *domain.DNSRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *DNSRecordRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.DNSRecord{}, "id = ?", id).Error
}

func (r *DNSRecordRepositoryImpl) DeleteByDomainID(ctx context.Context, domainID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.DNSRecord{}, "domain_id = ?", domainID).Error
}

func (r *DNSRecordRepositoryImpl) BulkCreate(ctx context.Context, records []*domain.DNSRecord) error {
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

func (r *DNSRecordRepositoryImpl) BulkUpdate(ctx context.Context, records []*domain.DNSRecord) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, record := range records {
			if err := tx.Save(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *DNSRecordRepositoryImpl) GetManagedRecords(ctx context.Context, domainID uuid.UUID) ([]*domain.DNSRecord, error) {
	var records []*domain.DNSRecord
	err := r.db.WithContext(ctx).Where("domain_id = ? AND is_managed = ?", domainID, true).Find(&records).Error
	return records, err
}

func (r *DNSRecordRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&domain.DNSRecord{}).Where("id = ?", id).Update("status", status).Error
}

func (r *DNSRecordRepositoryImpl) GetStaleRecords(ctx context.Context, hours int) ([]*domain.DNSRecord, error) {
	staleTime := time.Now().Add(time.Duration(-hours) * time.Hour)
	var records []*domain.DNSRecord
	err := r.db.WithContext(ctx).Where("last_checked_at < ? OR last_checked_at IS NULL", staleTime).Find(&records).Error
	return records, err
}
