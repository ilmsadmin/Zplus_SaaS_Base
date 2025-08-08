package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// FileServiceImpl implements FileService
type FileServiceImpl struct {
	fileRepo     domain.FileRepository
	auditService AuditService
	storageDir   string
	baseURL      string
}

// NewFileService creates a new file service
func NewFileService(
	fileRepo domain.FileRepository,
	auditService AuditService,
	storageDir string,
	baseURL string,
) FileService {
	return &FileServiceImpl{
		fileRepo:     fileRepo,
		auditService: auditService,
		storageDir:   storageDir,
		baseURL:      baseURL,
	}
}

// UploadFile uploads a file and stores metadata
func (s *FileServiceImpl) UploadFile(ctx context.Context, req *FileUploadRequest) (*FileResponse, error) {
	if err := s.validateUploadRequest(req); err != nil {
		return nil, err
	}

	// Generate unique filename
	fileExt := filepath.Ext(req.File.Header.Filename)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), fileExt)

	// Create directory structure
	var relativePath string
	if req.TenantID != nil {
		relativePath = filepath.Join("tenants", req.TenantID.String(), req.Category, fileName)
	} else {
		relativePath = filepath.Join("system", req.Category, fileName)
	}

	fullPath := filepath.Join(s.storageDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Save file to disk
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, req.File.File); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create file record
	file := &domain.File{
		ID:           uuid.New(),
		TenantID:     req.TenantID,
		UserID:       req.UserID,
		FileName:     fileName,
		OriginalName: req.File.Header.Filename,
		MimeType:     req.File.MimeType,
		Size:         req.File.Size,
		FilePath:     relativePath,
		URL:          fmt.Sprintf("%s/files/%s", s.baseURL, relativePath),
		StorageType:  "local",
		Category:     req.Category,
		Tags:         req.Tags,
		IsPublic:     req.IsPublic,
		ExpiryDate:   req.ExpiresAt,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		// Clean up file if database insert fails
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		tenantID := uuid.Nil
		if req.TenantID != nil {
			tenantID = *req.TenantID
		}
		s.auditService.LogEvent(ctx, tenantID, &req.UserID, domain.ActionCreate, domain.ResourceFile, file.ID.String(), map[string]interface{}{
			"file_name": file.OriginalName,
			"category":  file.Category,
			"size":      file.Size,
		})
	}

	return s.toFileResponse(file), nil
}

// DeleteFile deletes a file and its metadata
func (s *FileServiceImpl) DeleteFile(ctx context.Context, fileID uuid.UUID) error {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Delete physical file
	fullPath := filepath.Join(s.storageDir, file.FilePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete physical file: %w", err)
	}

	// Soft delete from database
	if err := s.fileRepo.SoftDelete(ctx, fileID); err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		tenantID := uuid.Nil
		if file.TenantID != nil {
			tenantID = *file.TenantID
		}
		s.auditService.LogEvent(ctx, tenantID, &file.UserID, domain.ActionDelete, domain.ResourceFile, file.ID.String(), map[string]interface{}{
			"file_name": file.OriginalName,
			"category":  file.Category,
		})
	}

	return nil
}

// GetFile retrieves file metadata
func (s *FileServiceImpl) GetFile(ctx context.Context, fileID uuid.UUID) (*FileResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return s.toFileResponse(file), nil
}

// GetFileByPath retrieves file by path
func (s *FileServiceImpl) GetFileByPath(ctx context.Context, path string) (*FileResponse, error) {
	file, err := s.fileRepo.GetByPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file by path: %w", err)
	}
	return s.toFileResponse(file), nil
}

// Helper methods

func (s *FileServiceImpl) validateUploadRequest(req *FileUploadRequest) error {
	if req.File == nil {
		return errors.New("file is required")
	}

	if req.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}

	// Check file size
	if req.MaxSize > 0 && req.File.Size > req.MaxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", req.File.Size, req.MaxSize)
	}

	// Check file type
	if len(req.AllowedTypes) > 0 {
		allowed := false
		for _, allowedType := range req.AllowedTypes {
			if req.File.MimeType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file type %s is not allowed", req.File.MimeType)
		}
	}

	// Validate file extension
	ext := filepath.Ext(req.File.Header.Filename)
	if ext == "" {
		return errors.New("file must have an extension")
	}

	// Check for dangerous extensions
	dangerousExts := []string{".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js", ".jar", ".php", ".asp", ".jsp"}
	for _, dangerousExt := range dangerousExts {
		if strings.EqualFold(ext, dangerousExt) {
			return fmt.Errorf("file type %s is not allowed for security reasons", ext)
		}
	}

	return nil
}

func (s *FileServiceImpl) toFileResponse(file *domain.File) *FileResponse {
	return &FileResponse{
		ID:           file.ID,
		TenantID:     file.TenantID,
		UserID:       file.UserID,
		FileName:     file.FileName,
		OriginalName: file.OriginalName,
		MimeType:     file.MimeType,
		Size:         file.Size,
		Path:         file.FilePath,
		URL:          file.URL,
		StorageType:  file.StorageType,
		Category:     file.Category,
		Tags:         file.Tags,
		IsPublic:     file.IsPublic,
		ExpiresAt:    file.ExpiryDate,
		Metadata:     file.Metadata,
		CreatedAt:    file.CreatedAt,
		UpdatedAt:    file.UpdatedAt,
	}
}
