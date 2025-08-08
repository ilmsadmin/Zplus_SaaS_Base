package application

import (
	"time"

	"github.com/google/uuid"
)

// UserBasicDTO represents basic user information
type UserBasicDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Avatar    *string   `json:"avatar"`
	AvatarURL *string   `json:"avatar_url"`
}

// FileUploadDTO represents file upload request data
type FileUploadDTO struct {
	FileName    string                 `json:"file_name" validate:"required"`
	FileSize    int64                  `json:"file_size" validate:"required,min=1"`
	MimeType    string                 `json:"mime_type" validate:"required"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	IsPublic    bool                   `json:"is_public"`
	ExpiryDate  *time.Time             `json:"expiry_date"`
	Metadata    map[string]interface{} `json:"metadata"`
	ChunkSize   int                    `json:"chunk_size"` // For chunked uploads
	TotalChunks int                    `json:"total_chunks"`
}

// FileUploadSessionResponseDTO represents upload session response
type FileUploadSessionResponseDTO struct {
	SessionToken   string    `json:"session_token"`
	UploadURL      string    `json:"upload_url"`
	ChunkSize      int       `json:"chunk_size"`
	TotalChunks    int       `json:"total_chunks"`
	UploadedChunks int       `json:"uploaded_chunks"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// FileUploadProgressDTO represents upload progress
type FileUploadProgressDTO struct {
	SessionToken      string  `json:"session_token"`
	UploadedChunks    int     `json:"uploaded_chunks"`
	TotalChunks       int     `json:"total_chunks"`
	ProgressPercent   float64 `json:"progress_percent"`
	UploadedBytes     int64   `json:"uploaded_bytes"`
	TotalBytes        int64   `json:"total_bytes"`
	EstimatedTimeLeft int64   `json:"estimated_time_left"` // seconds
	UploadSpeed       float64 `json:"upload_speed"`        // bytes per second
}

// ChunkUploadDTO represents a file chunk upload
type ChunkUploadDTO struct {
	SessionToken string `json:"session_token" validate:"required"`
	ChunkNumber  int    `json:"chunk_number" validate:"required,min=0"`
	ChunkData    []byte `json:"chunk_data" validate:"required"`
	ChunkHash    string `json:"chunk_hash"` // For verification
}

// FileResponseDTO represents file response data
type FileResponseDTO struct {
	ID               uuid.UUID              `json:"id"`
	FileName         string                 `json:"file_name"`
	OriginalName     string                 `json:"original_name"`
	MimeType         string                 `json:"mime_type"`
	Size             int64                  `json:"size"`
	URL              string                 `json:"url"`
	DownloadURL      string                 `json:"download_url"`
	PreviewURL       string                 `json:"preview_url,omitempty"`
	ThumbnailURL     string                 `json:"thumbnail_url,omitempty"`
	Category         string                 `json:"category"`
	Tags             []string               `json:"tags"`
	IsPublic         bool                   `json:"is_public"`
	IsProcessed      bool                   `json:"is_processed"`
	ProcessingStatus string                 `json:"processing_status"`
	VirusScanStatus  string                 `json:"virus_scan_status"`
	ExpiryDate       *time.Time             `json:"expiry_date"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	User             *UserBasicDTO          `json:"user,omitempty"`
}

// FileListResponseDTO represents paginated file list response
type FileListResponseDTO struct {
	Files      []*FileResponseDTO `json:"files"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
	TotalSize  int64              `json:"total_size"` // Total size of all files
}

// FileUpdateDTO represents file update request
type FileUpdateDTO struct {
	FileName   *string                `json:"file_name,omitempty"`
	Category   *string                `json:"category,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	IsPublic   *bool                  `json:"is_public,omitempty"`
	ExpiryDate *time.Time             `json:"expiry_date,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// FileSearchDTO represents file search criteria
type FileSearchDTO struct {
	Query     string    `json:"query"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	MimeType  string    `json:"mime_type"`
	MinSize   int64     `json:"min_size"`
	MaxSize   int64     `json:"max_size"`
	IsPublic  *bool     `json:"is_public"`
	DateFrom  time.Time `json:"date_from"`
	DateTo    time.Time `json:"date_to"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
	SortBy    string    `json:"sort_by"`    // name, size, date, category
	SortOrder string    `json:"sort_order"` // asc, desc
	OnlyMine  bool      `json:"only_mine"`  // Only current user's files
}

// FileShareCreateDTO represents file share creation request
type FileShareCreateDTO struct {
	FileID       uuid.UUID  `json:"file_id" validate:"required"`
	SharedWith   *uuid.UUID `json:"shared_with,omitempty"` // NULL for public share
	ShareType    string     `json:"share_type" validate:"required,oneof=read write download"`
	Password     string     `json:"password,omitempty"`
	MaxDownloads int        `json:"max_downloads"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

// FileShareResponseDTO represents file share response
type FileShareResponseDTO struct {
	ID            uuid.UUID        `json:"id"`
	FileID        uuid.UUID        `json:"file_id"`
	AccessToken   string           `json:"access_token"`
	ShareType     string           `json:"share_type"`
	ShareURL      string           `json:"share_url"`
	MaxDownloads  int              `json:"max_downloads"`
	DownloadCount int              `json:"download_count"`
	ExpiresAt     *time.Time       `json:"expires_at"`
	IsActive      bool             `json:"is_active"`
	CreatedAt     time.Time        `json:"created_at"`
	File          *FileResponseDTO `json:"file,omitempty"`
	SharedBy      *UserBasicDTO    `json:"shared_by,omitempty"`
	SharedWith    *UserBasicDTO    `json:"shared_with,omitempty"`
}

// FileShareAccessDTO represents file share access request
type FileShareAccessDTO struct {
	AccessToken string `json:"access_token" validate:"required"`
	Password    string `json:"password,omitempty"`
}

// ImageProcessingDTO represents image processing request
type ImageProcessingDTO struct {
	FileID     uuid.UUID              `json:"file_id" validate:"required"`
	Operations []ImageOperationDTO    `json:"operations" validate:"required,min=1"`
	OutputName string                 `json:"output_name,omitempty"`
	Quality    int                    `json:"quality,omitempty"` // 1-100
	Format     string                 `json:"format,omitempty"`  // jpeg, png, webp
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ImageOperationDTO represents individual image operation
type ImageOperationDTO struct {
	Type   string                 `json:"type" validate:"required,oneof=resize crop rotate compress thumbnail"`
	Params map[string]interface{} `json:"params" validate:"required"`
}

// FileProcessingJobResponseDTO represents processing job response
type FileProcessingJobResponseDTO struct {
	ID           uuid.UUID              `json:"id"`
	FileID       uuid.UUID              `json:"file_id"`
	JobType      string                 `json:"job_type"`
	Status       string                 `json:"status"`
	Progress     int                    `json:"progress"`
	Result       map[string]interface{} `json:"result"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
}

// FileStorageConfigDTO represents storage configuration
type FileStorageConfigDTO struct {
	ID               uuid.UUID              `json:"id"`
	StorageType      string                 `json:"storage_type"`
	Config           map[string]interface{} `json:"config"`
	IsActive         bool                   `json:"is_active"`
	MaxFileSize      int64                  `json:"max_file_size"`
	AllowedMimeTypes []string               `json:"allowed_mime_types"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// FileStorageConfigCreateDTO represents storage configuration creation
type FileStorageConfigCreateDTO struct {
	StorageType      string                 `json:"storage_type" validate:"required,oneof=local s3 minio azure gcs"`
	Config           map[string]interface{} `json:"config" validate:"required"`
	MaxFileSize      int64                  `json:"max_file_size" validate:"min=1"`
	AllowedMimeTypes []string               `json:"allowed_mime_types" validate:"required,min=1"`
}

// FileStorageConfigUpdateDTO represents storage configuration update
type FileStorageConfigUpdateDTO struct {
	Config           map[string]interface{} `json:"config,omitempty"`
	MaxFileSize      *int64                 `json:"max_file_size,omitempty"`
	AllowedMimeTypes []string               `json:"allowed_mime_types,omitempty"`
	IsActive         *bool                  `json:"is_active,omitempty"`
}

// FileStatsDTO represents file statistics
type FileStatsDTO struct {
	TotalFiles      int64                       `json:"total_files"`
	TotalSize       int64                       `json:"total_size"`
	TotalSizeHuman  string                      `json:"total_size_human"`
	FilesByCategory map[string]int64            `json:"files_by_category"`
	FilesByMimeType map[string]int64            `json:"files_by_mime_type"`
	FilesByMonth    map[string]int64            `json:"files_by_month"`
	StorageUsage    map[string]int64            `json:"storage_usage"`
	RecentActivity  []*FileAccessLogResponseDTO `json:"recent_activity"`
}

// FileAccessLogResponseDTO represents file access log
type FileAccessLogResponseDTO struct {
	ID        uuid.UUID        `json:"id"`
	FileID    uuid.UUID        `json:"file_id"`
	Action    string           `json:"action"`
	IPAddress string           `json:"ip_address"`
	UserAgent string           `json:"user_agent"`
	CreatedAt time.Time        `json:"created_at"`
	User      *UserBasicDTO    `json:"user,omitempty"`
	File      *FileResponseDTO `json:"file,omitempty"`
}

// VirusScanResultDTO represents virus scan result
type VirusScanResultDTO struct {
	FileID      uuid.UUID              `json:"file_id"`
	Status      string                 `json:"status"` // clean, infected, error
	ScannerName string                 `json:"scanner_name"`
	ScanDate    time.Time              `json:"scan_date"`
	Threats     []string               `json:"threats,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// FileCleanupDTO represents file cleanup configuration
type FileCleanupDTO struct {
	DeleteExpired     bool      `json:"delete_expired"`
	DeleteUnused      bool      `json:"delete_unused"`
	DeleteOlderThan   time.Time `json:"delete_older_than"`
	DryRun            bool      `json:"dry_run"`
	Categories        []string  `json:"categories,omitempty"`
	ExcludeCategories []string  `json:"exclude_categories,omitempty"`
}

// FileCleanupResultDTO represents cleanup operation result
type FileCleanupResultDTO struct {
	DeletedFiles     int64  `json:"deleted_files"`
	DeletedSize      int64  `json:"deleted_size"`
	DeletedSizeHuman string `json:"deleted_size_human"`
	ErrorCount       int    `json:"error_count"`
	Duration         int64  `json:"duration_ms"`
}
