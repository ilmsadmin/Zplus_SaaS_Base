package infrastructure

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"gorm.io/gorm"
)

// FileRepository implements domain.FileRepository
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create creates a new file record
func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// GetByID retrieves a file by ID
func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&file).Error

	if err != nil {
		return nil, err
	}

	return &file, nil
}

// GetByPath retrieves a file by path
func (r *FileRepository) GetByPath(ctx context.Context, path string) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).
		Where("file_path = ? AND deleted_at IS NULL", path).
		First(&file).Error

	if err != nil {
		return nil, err
	}

	return &file, nil
}

// GetByChecksum retrieves a file by checksum
func (r *FileRepository) GetByChecksum(ctx context.Context, checksum string) (*domain.File, error) {
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
func (r *FileRepository) Update(ctx context.Context, file *domain.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// Delete permanently deletes a file record
func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.File{}).Error
}

// SoftDelete soft deletes a file record
func (r *FileRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.File{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

// ListByUser lists files by user ID with pagination
func (r *FileRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// ListByTenant lists files by tenant ID with pagination
func (r *FileRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// ListByCategory lists files by category with pagination
func (r *FileRepository) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("category = ? AND deleted_at IS NULL", category).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// ListByUserAndCategory lists files by user and category with pagination
func (r *FileRepository) ListByUserAndCategory(ctx context.Context, userID uuid.UUID, category string, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("user_id = ? AND category = ? AND deleted_at IS NULL", userID, category).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// ListByTags lists files by tags with pagination
func (r *FileRepository) ListByTags(ctx context.Context, tags []string, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("deleted_at IS NULL")

	for _, tag := range tags {
		query = query.Where("? = ANY(tags)", tag)
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// ListPublic lists public files with pagination
func (r *FileRepository) ListPublic(ctx context.Context, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("is_public = ? AND deleted_at IS NULL", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// ListPendingProcessing lists files pending processing
func (r *FileRepository) ListPendingProcessing(ctx context.Context, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("processing_status = ? AND deleted_at IS NULL", "pending").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// ListPendingVirusScan lists files pending virus scan
func (r *FileRepository) ListPendingVirusScan(ctx context.Context, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File
	err := r.db.WithContext(ctx).
		Where("virus_scan_status = ? AND deleted_at IS NULL", "pending").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// CountByUser counts files by user ID
func (r *FileRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.File{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error

	return count, err
}

// CountByTenant counts files by tenant ID
func (r *FileRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.File{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error

	return count, err
}

// GetTotalSizeByUser gets total file size by user ID
func (r *FileRepository) GetTotalSizeByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var totalSize int64
	err := r.db.WithContext(ctx).Model(&domain.File{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error

	return totalSize, err
}

// GetTotalSizeByTenant gets total file size by tenant ID
func (r *FileRepository) GetTotalSizeByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var totalSize int64
	err := r.db.WithContext(ctx).Model(&domain.File{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error

	return totalSize, err
}

// DeleteExpired deletes expired files
func (r *FileRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&domain.File{}).
		Where("expiry_date IS NOT NULL AND expiry_date < ? AND deleted_at IS NULL", time.Now()).
		Update("deleted_at", time.Now()).Error
}

// Search searches files by query with pagination
func (r *FileRepository) Search(ctx context.Context, query string, tenantID *uuid.UUID, limit, offset int) ([]*domain.File, error) {
	var files []*domain.File

	dbQuery := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("deleted_at IS NULL")

	if tenantID != nil {
		dbQuery = dbQuery.Where("tenant_id = ?", *tenantID)
	}

	// Search in filename, original name, and tags
	searchPattern := "%" + strings.ToLower(query) + "%"
	dbQuery = dbQuery.Where(
		"LOWER(file_name) LIKE ? OR LOWER(original_name) LIKE ? OR EXISTS (SELECT 1 FROM unnest(tags) AS tag WHERE LOWER(tag) LIKE ?)",
		searchPattern, searchPattern, searchPattern,
	)

	err := dbQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	return files, err
}

// UpdateProcessingStatus updates file processing status
func (r *FileRepository) UpdateProcessingStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&domain.File{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"processing_status": status,
			"is_processed":      status == "completed",
		}).Error
}

// UpdateVirusScanStatus updates file virus scan status
func (r *FileRepository) UpdateVirusScanStatus(ctx context.Context, id uuid.UUID, status string, result map[string]interface{}) error {
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

// FileStorageConfigRepository implements domain.FileStorageConfigRepository
type FileStorageConfigRepository struct {
	db *gorm.DB
}

// NewFileStorageConfigRepository creates a new file storage config repository
func NewFileStorageConfigRepository(db *gorm.DB) *FileStorageConfigRepository {
	return &FileStorageConfigRepository{db: db}
}

// Create creates a new storage configuration
func (r *FileStorageConfigRepository) Create(ctx context.Context, config *domain.FileStorageConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetByID retrieves a storage configuration by ID
func (r *FileStorageConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.FileStorageConfig, error) {
	var config domain.FileStorageConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByTenantAndType retrieves a storage configuration by tenant and type
func (r *FileStorageConfigRepository) GetByTenantAndType(ctx context.Context, tenantID uuid.UUID, storageType string) (*domain.FileStorageConfig, error) {
	var config domain.FileStorageConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND storage_type = ?", tenantID, storageType).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetActiveByTenant retrieves the active storage configuration for a tenant
func (r *FileStorageConfigRepository) GetActiveByTenant(ctx context.Context, tenantID uuid.UUID) (*domain.FileStorageConfig, error) {
	var config domain.FileStorageConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Update updates a storage configuration
func (r *FileStorageConfigRepository) Update(ctx context.Context, config *domain.FileStorageConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete deletes a storage configuration
func (r *FileStorageConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.FileStorageConfig{}).Error
}

// ListByTenant lists storage configurations by tenant
func (r *FileStorageConfigRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.FileStorageConfig, error) {
	var configs []*domain.FileStorageConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&configs).Error
	return configs, err
}

// SetActive sets a storage configuration as active and deactivates others
func (r *FileStorageConfigRepository) SetActive(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate all configs for the tenant
		if err := tx.Model(&domain.FileStorageConfig{}).
			Where("tenant_id = ?", tenantID).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the specified config
		return tx.Model(&domain.FileStorageConfig{}).
			Where("id = ? AND tenant_id = ?", id, tenantID).
			Update("is_active", true).Error
	})
}
