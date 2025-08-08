package infrastructure

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorageProvider implements local file storage
type LocalStorageProvider struct {
	basePath  string
	publicURL string
}

// NewLocalStorageProvider creates a new local storage provider
func NewLocalStorageProvider(basePath, publicURL string) *LocalStorageProvider {
	return &LocalStorageProvider{
		basePath:  basePath,
		publicURL: publicURL,
	}
}

// Store stores a file locally
func (p *LocalStorageProvider) Store(ctx context.Context, path string, data io.Reader, metadata map[string]string) error {
	fullPath := filepath.Join(p.basePath, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	if _, err := io.Copy(file, data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Retrieve retrieves a file from local storage
func (p *LocalStorageProvider) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(p.basePath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete deletes a file from local storage
func (p *LocalStorageProvider) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(p.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL returns the public URL for a file
func (p *LocalStorageProvider) GetURL(ctx context.Context, path string) (string, error) {
	return fmt.Sprintf("%s/%s", p.publicURL, path), nil
}

// GetSignedURL returns a signed URL for temporary access (same as GetURL for local storage)
func (p *LocalStorageProvider) GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	// For local storage, we return the same URL as GetURL
	// In a real implementation, you might want to implement time-based tokens
	return p.GetURL(ctx, path)
}
