package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"gorm.io/gorm"
)

// FileRepositoryImpl implements domain.FileRepository
type FileRepositoryImpl struct {
	db *gorm.DB
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *gorm.DB) domain.FileRepository {
	return &FileRepositoryImpl{db: db}
}

// Create creates a new file record
func (r *FileRepositoryImpl) Create(ctx context.Context, file *domain.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// GetByID retrieves a file by ID
func (r *FileRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByPath retrieves a file by path
func (r *FileRepositoryImpl) GetByPath(ctx context.Context, path string) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		First(&file, "path = ?", path).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByChecksum retrieves a file by checksum
func (r *FileRepositoryImpl) GetByChecksum(ctx context.Context, checksum string) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).
		Where("checksum = ? AND deleted_at IS NULL", checksum).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// Update updates a file record
func (r *FileRepositoryImpl) Update(ctx context.Context, file *domain.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// Delete hard deletes a file record
func (r *FileRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.File{}, "id = ?", id).Error
}

// SoftDelete soft deletes a file record
func (r *FileRepositoryImpl) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.File{}, "id = ?", id).Error
}

// ListByUser retrieves files by user with pagination
func (r *FileRepositoryImpl) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListByTenant retrieves files by tenant with pagination
func (r *FileRepositoryImpl) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListByCategory retrieves files by category with pagination
func (r *FileRepositoryImpl) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("category = ?", category).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListByUserAndCategory retrieves files by user and category with pagination
func (r *FileRepositoryImpl) ListByUserAndCategory(ctx context.Context, userID uuid.UUID, category string, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND category = ?", userID, category).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListByTags lists files by tags
func (r *FileRepositoryImpl) ListByTags(ctx context.Context, tags []string, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("tags @> ?", tags).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListPublic lists public files
func (r *FileRepositoryImpl) ListPublic(ctx context.Context, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListPendingProcessing lists files pending processing
func (r *FileRepositoryImpl) ListPendingProcessing(ctx context.Context, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("processing_status = ?", "pending").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListPendingVirusScan lists files pending virus scan
func (r *FileRepositoryImpl) ListPendingVirusScan(ctx context.Context, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("virus_scan_status = ?", "pending").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// CountByUser counts files by user
func (r *FileRepositoryImpl) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByTenant counts files by tenant
func (r *FileRepositoryImpl) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

// GetTotalSizeByUser gets total file size by user
func (r *FileRepositoryImpl) GetTotalSizeByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var totalSize int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error
	return totalSize, err
}

// GetTotalSizeByTenant gets total file size by tenant
func (r *FileRepositoryImpl) GetTotalSizeByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var totalSize int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error
	return totalSize, err
}

// DeleteExpired deletes expired files
func (r *FileRepositoryImpl) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Delete(&domain.File{}, "expires_at IS NOT NULL AND expires_at < ?", time.Now()).Error
}

// Search searches files by name and content
func (r *FileRepositoryImpl) Search(ctx context.Context, query string, tenantID *uuid.UUID, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	searchPattern := "%" + query + "%"

	db := r.db.WithContext(ctx).
		Where("file_name ILIKE ? OR original_name ILIKE ?", searchPattern, searchPattern)

	if tenantID != nil {
		db = db.Where("tenant_id = ?", *tenantID)
	}

	err := db.Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// UpdateProcessingStatus updates file processing status
func (r *FileRepositoryImpl) UpdateProcessingStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("id = ?", id).
		Update("processing_status", status).Error
}

// UpdateVirusScanStatus updates file virus scan status
func (r *FileRepositoryImpl) UpdateVirusScanStatus(ctx context.Context, id uuid.UUID, status string, result map[string]interface{}) error {
	updates := map[string]interface{}{
		"virus_scan_status": status,
	}

	if result != nil {
		updates["virus_scan_result"] = result
	}

	return r.db.WithContext(ctx).Model(&domain.File{}).
		Where("id = ?", id).
		Updates(updates).Error
}
