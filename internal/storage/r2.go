package storage

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/walle233/yt-downloader/internal/config"
)

type R2Storage struct {
	bucket   string
	public   string
	client   *s3.Client
	presign  *s3.PresignClient
	endpoint string
}

func NewR2Storage(ctx context.Context, cfg config.R2Config) (*R2Storage, error) {
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.Bucket == "" || cfg.Endpoint == "" {
		return nil, fmt.Errorf("missing required R2 configuration")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = awsv2.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &R2Storage{
		bucket:   cfg.Bucket,
		public:   strings.TrimRight(cfg.PublicBaseURL, "/"),
		client:   client,
		presign:  s3.NewPresignClient(client),
		endpoint: cfg.Endpoint,
	}, nil
}

func (s *R2Storage) UploadFile(ctx context.Context, objectKey, filePath string) error {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      awsv2.String(s.bucket),
		Key:         awsv2.String(objectKey),
		Body:        file,
		ContentType: awsv2.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}

func (s *R2Storage) PresignGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	if s.public != "" && !strings.Contains(s.public, ".r2.cloudflarestorage.com") {
		return s.public + "/" + strings.TrimLeft(objectKey, "/"), nil
	}

	result, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: awsv2.String(s.bucket),
		Key:    awsv2.String(objectKey),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}

	return result.URL, nil
}
