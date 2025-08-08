package infrastructure

import (
	"context"
	"io"
	"time"
)

// ImageResizeParams contains parameters for image resizing
type ImageResizeParams struct {
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Quality int    `json:"quality"`
	Format  string `json:"format"`
}

// ImageCropParams contains parameters for image cropping
type ImageCropParams struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Quality int    `json:"quality"`
	Format  string `json:"format"`
}

// ImageInfo contains image metadata
type ImageInfo struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Format   string `json:"format"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// VirusScanResult contains virus scan results
type VirusScanResult struct {
	Safe      bool          `json:"safe"`
	Threats   []string      `json:"threats,omitempty"`
	ScanTime  time.Duration `json:"scan_time"`
	Scanner   string        `json:"scanner"`
	ScannedAt time.Time     `json:"scanned_at"`
}

// Type aliases for infrastructure components
type (
	StorageProvider = StorageProviderInterface
	ImageProcessor  = ImageProcessorInterface
	VirusScanner    = VirusScannerInterface
)

// StorageProviderInterface defines file storage operations
type StorageProviderInterface interface {
	Upload(ctx context.Context, path string, reader io.Reader) error
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	GetURL(ctx context.Context, path string) (string, error)
	GetSize(ctx context.Context, path string) (int64, error)
}

// ImageProcessorInterface defines image processing operations
type ImageProcessorInterface interface {
	Resize(ctx context.Context, reader io.Reader, params ImageResizeParams) (io.ReadCloser, error)
	Crop(ctx context.Context, reader io.Reader, params ImageCropParams) (io.ReadCloser, error)
	GenerateThumbnail(ctx context.Context, reader io.Reader, width, height int) (io.ReadCloser, error)
	GetImageInfo(ctx context.Context, reader io.Reader) (*ImageInfo, error)
}

// VirusScannerInterface defines virus scanning operations
type VirusScannerInterface interface {
	ScanFile(ctx context.Context, reader io.Reader, filename string) (*VirusScanResult, error)
	IsEnabled() bool
	GetScannerName() string
}
