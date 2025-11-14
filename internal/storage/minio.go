package storage

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	"inteam/internal/config"
)

type ObjectStorage interface {
	SaveProfileSnapshot(ctx context.Context, vkID int64, raw []byte) error
}

type minioStorage struct {
	client *minio.Client
	bucket string
	logger *zap.Logger
}

func NewMinio(cfg config.MinioConfig, logger *zap.Logger) (ObjectStorage, error) {
	if cfg.Endpoint == "" {
		logger.Info("minio disabled: endpoint is empty")
		return nil, nil
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	logger.Info("minio connected", zap.String("endpoint", cfg.Endpoint), zap.String("bucket", cfg.Bucket))

	return &minioStorage{
		client: client,
		bucket: cfg.Bucket,
		logger: logger,
	}, nil
}

func (s *minioStorage) SaveProfileSnapshot(ctx context.Context, vkID int64, raw []byte) error {
	objectName := fmt.Sprintf("profiles/%s.json", strconv.FormatInt(vkID, 10))

	_, err := s.client.PutObject(ctx, s.bucket, objectName, bytes.NewReader(raw), int64(len(raw)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		s.logger.Warn("failed to save profile snapshot", zap.Error(err))
		return err
	}

	return nil
}

