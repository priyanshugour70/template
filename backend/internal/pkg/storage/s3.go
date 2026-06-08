// Package storage wraps S3 for presigned-URL upload/download. Server signs;
// clients PUT/GET bytes direct to S3. file_uploads table tracks the lifecycle.
package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/your-org/your-service/internal/config"
)

type S3 struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucket     string
	keyPrefix  string
	publicBase string
	presignTTL time.Duration
}

// NewS3 builds a client from config.Assets. Returns nil, nil if S3 is not
// configured (so callers can run in dev without AWS).
func NewS3(ctx context.Context, cfg config.Assets) (*S3, error) {
	if strings.TrimSpace(cfg.S3Region) == "" || strings.TrimSpace(cfg.S3Bucket) == "" {
		return nil, nil
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.S3Region))
	if err != nil {
		return nil, fmt.Errorf("s3: load aws config: %w", err)
	}
	cl := s3.NewFromConfig(awsCfg)
	ttl := time.Duration(cfg.PresignTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &S3{
		client:     cl,
		presigner:  s3.NewPresignClient(cl),
		bucket:     cfg.S3Bucket,
		keyPrefix:  strings.Trim(cfg.S3KeyPrefix, "/"),
		publicBase: strings.TrimRight(cfg.S3PublicBaseURL, "/"),
		presignTTL: ttl,
	}, nil
}

func (s *S3) Bucket() string         { return s.bucket }
func (s *S3) PresignTTL() time.Duration { return s.presignTTL }

// Key joins the configured prefix with the given parts, normalising slashes.
func (s *S3) Key(parts ...string) string {
	out := s.keyPrefix
	for _, p := range parts {
		p = strings.Trim(p, "/")
		if p == "" {
			continue
		}
		if out == "" {
			out = p
		} else {
			out = out + "/" + p
		}
	}
	return out
}

func (s *S3) PublicURL(key string) string {
	if s.publicBase == "" {
		return ""
	}
	return s.publicBase + "/" + strings.TrimLeft(key, "/")
}

// PresignPut returns a URL that lets the holder PUT bytes directly. Caller
// must send the same Content-Type they signed for.
func (s *S3) PresignPut(ctx context.Context, key, contentType string) (url string, expiresAt time.Time, err error) {
	req, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("s3 presign put: %w", err)
	}
	return req.URL, time.Now().Add(s.presignTTL), nil
}

// PresignGet returns a URL that lets the holder GET an object.
func (s *S3) PresignGet(ctx context.Context, key string) (url string, expiresAt time.Time, err error) {
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("s3 presign get: %w", err)
	}
	return req.URL, time.Now().Add(s.presignTTL), nil
}

// Head returns object size, etag (no quotes), and content type.
func (s *S3) Head(ctx context.Context, key string) (size int64, etag, mime string, err error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, "", "", fmt.Errorf("s3 head: %w", err)
	}
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	if out.ETag != nil {
		etag = strings.Trim(*out.ETag, `"`)
	}
	if out.ContentType != nil {
		mime = *out.ContentType
	}
	return size, etag, mime, nil
}

// Delete removes an object. Use sparingly — prefer marking file_uploads.status.
func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}
	return nil
}
