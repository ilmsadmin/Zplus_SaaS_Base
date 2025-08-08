package infrastructure

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3StorageProvider implements S3-compatible storage
type S3StorageProvider struct {
	client     *s3.Client
	bucket     string
	region     string
	endpoint   string // For S3-compatible services like MinIO
	publicURL  string
	pathPrefix string
}

// S3Config represents S3 configuration
type S3Config struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint,omitempty"`         // For S3-compatible services
	PathPrefix      string `json:"path_prefix,omitempty"`      // Optional path prefix
	PublicURL       string `json:"public_url,omitempty"`       // Custom public URL base
	ForcePathStyle  bool   `json:"force_path_style,omitempty"` // For MinIO compatibility
}

// NewS3StorageProvider creates a new S3 storage provider
func NewS3StorageProvider(cfg S3Config) (*S3StorageProvider, error) {
	// Create AWS config
	var awsCfg aws.Config
	var err error

	if cfg.Endpoint != "" {
		// For S3-compatible services (MinIO, etc.)
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           cfg.Endpoint,
				SigningRegion: cfg.Region,
			}, nil
		})

		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     cfg.AccessKeyID,
					SecretAccessKey: cfg.SecretAccessKey,
				}, nil
			}))),
		)
	} else {
		// Standard AWS S3
		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.ForcePathStyle {
			o.UsePathStyle = true
		}
	})

	publicURL := cfg.PublicURL
	if publicURL == "" {
		if cfg.Endpoint != "" {
			publicURL = fmt.Sprintf("%s/%s", cfg.Endpoint, cfg.Bucket)
		} else {
			publicURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, cfg.Region)
		}
	}

	return &S3StorageProvider{
		client:     client,
		bucket:     cfg.Bucket,
		region:     cfg.Region,
		endpoint:   cfg.Endpoint,
		publicURL:  publicURL,
		pathPrefix: cfg.PathPrefix,
	}, nil
}

// Store stores a file in S3
func (p *S3StorageProvider) Store(ctx context.Context, path string, data io.Reader, metadata map[string]string) error {
	key := p.buildKey(path)

	// Convert metadata to S3 metadata
	s3Metadata := make(map[string]string)
	for k, v := range metadata {
		// S3 metadata keys must be lowercase and contain only letters, numbers, and hyphens
		cleanKey := strings.ToLower(strings.ReplaceAll(k, "_", "-"))
		s3Metadata[cleanKey] = v
	}

	// Upload object
	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &p.bucket,
		Key:         &key,
		Body:        data,
		Metadata:    s3Metadata,
		ContentType: aws.String(metadata["content-type"]),
	})

	if err != nil {
		return fmt.Errorf("failed to upload object to S3: %w", err)
	}

	return nil
}

// Retrieve retrieves a file from S3
func (p *S3StorageProvider) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	key := p.buildKey(path)

	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	return result.Body, nil
}

// Delete deletes a file from S3
func (p *S3StorageProvider) Delete(ctx context.Context, path string) error {
	key := p.buildKey(path)

	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})

	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}

// GetURL returns the public URL for a file
func (p *S3StorageProvider) GetURL(ctx context.Context, path string) (string, error) {
	key := p.buildKey(path)
	return fmt.Sprintf("%s/%s", p.publicURL, key), nil
}

// GetSignedURL returns a pre-signed URL for temporary access
func (p *S3StorageProvider) GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	key := p.buildKey(path)

	presignClient := s3.NewPresignClient(p.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// CreateBucket creates the S3 bucket if it doesn't exist
func (p *S3StorageProvider) CreateBucket(ctx context.Context) error {
	// Check if bucket exists
	_, err := p.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &p.bucket,
	})

	if err == nil {
		// Bucket already exists
		return nil
	}

	// Create bucket
	createInput := &s3.CreateBucketInput{
		Bucket: &p.bucket,
	}

	// Set location constraint for regions other than us-east-1
	if p.region != "us-east-1" {
		createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(p.region),
		}
	}

	_, err = p.client.CreateBucket(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// SetBucketPolicy sets the bucket policy for public read access (optional)
func (p *S3StorageProvider) SetBucketPolicy(ctx context.Context, allowPublicRead bool) error {
	if !allowPublicRead {
		return nil
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "PublicReadGetObject",
				"Effect": "Allow",
				"Principal": "*",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/*"
			}
		]
	}`, p.bucket)

	_, err := p.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: &p.bucket,
		Policy: &policy,
	})

	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	return nil
}

// ListObjects lists objects in the bucket with a prefix
func (p *S3StorageProvider) ListObjects(ctx context.Context, prefix string, maxKeys int32) ([]string, error) {
	key := p.buildKey(prefix)

	result, err := p.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  &p.bucket,
		Prefix:  &key,
		MaxKeys: &maxKeys,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var objects []string
	for _, obj := range result.Contents {
		if obj.Key != nil {
			// Remove path prefix if it exists
			objKey := *obj.Key
			if p.pathPrefix != "" && strings.HasPrefix(objKey, p.pathPrefix+"/") {
				objKey = strings.TrimPrefix(objKey, p.pathPrefix+"/")
			}
			objects = append(objects, objKey)
		}
	}

	return objects, nil
}

// GetObjectInfo returns object metadata
func (p *S3StorageProvider) GetObjectInfo(ctx context.Context, path string) (*ObjectInfo, error) {
	key := p.buildKey(path)

	result, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	var contentType string
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	var size int64
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	var lastModified time.Time
	if result.LastModified != nil {
		lastModified = *result.LastModified
	}

	return &ObjectInfo{
		Key:          key,
		Size:         size,
		ContentType:  contentType,
		LastModified: lastModified,
		Metadata:     result.Metadata,
	}, nil
}

// ObjectInfo represents object metadata
type ObjectInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	Metadata     map[string]string `json:"metadata"`
}

// buildKey constructs the full S3 key with optional prefix
func (p *S3StorageProvider) buildKey(path string) string {
	if p.pathPrefix == "" {
		return path
	}
	return fmt.Sprintf("%s/%s", p.pathPrefix, path)
}
