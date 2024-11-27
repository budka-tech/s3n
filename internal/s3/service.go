package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	logit "github.com/budka-tech/logit-go"
	"github.com/google/uuid"
	"io"
	cfg "s3n/internal/config"
)

type S3Service struct {
	logger logit.Logger

	redirectFormat string
	fileFormat     string
	client         *s3.Client
	uploader       *manager.Uploader
}

func NewS3Service(ctx context.Context, logger logit.Logger, s3Config *cfg.S3ServiceConfig) (Service, error) {
	const op = "S3Service.NewS3Service"
	ctx = logger.NewOpCtx(ctx, op)

	clientCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3Config.S3Server.Region),
		config.WithBaseEndpoint(s3Config.S3Server.Endpoint),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     s3Config.S3Server.AccessKey,
				SecretAccessKey: s3Config.S3Server.SecretKey,
			},
		}),
	)
	if err != nil {
		err = fmt.Errorf("не удалось загрузить конфиг: %w", err)
		logger.Error(ctx, err)
		return nil, err
	}

	client := s3.NewFromConfig(clientCfg)

	uploader := manager.NewUploader(client, func(uploader *manager.Uploader) {
		uploader.Concurrency = s3Config.UploadGoroutines
		uploader.PartSize = s3Config.UploadPartSize
	})

	return &S3Service{
		logger:         logger,
		client:         client,
		uploader:       uploader,
		redirectFormat: s3Config.RedirectFormat,
		fileFormat:     s3Config.FileFormat,
	}, nil
}

func (s *S3Service) UploadFile(ctx context.Context, bucket string, key string, file io.Reader) error {
	const op = "S3Service.UploadFile"
	ctx = s.logger.NewOpCtx(ctx, op)

	// Upload the file
	_, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
		ACL:    types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		err = fmt.Errorf("не удалось загрузить файл: %w", err)
		s.logger.Error(ctx, err)
		return err
	}

	// Successfully uploaded
	return nil
}

func (s *S3Service) UploadFileBytes(ctx context.Context, bucket string, key string, file []byte) error {
	return s.UploadFile(ctx, bucket, key, bytes.NewReader(file))
}

func (s *S3Service) DeleteFile(ctx context.Context, bucket string, key string) error {
	const op = "S3Service.DeleteFile"
	ctx = s.logger.NewOpCtx(ctx, op)

	// Delete the object from S3
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		err = fmt.Errorf("не удалось удалить файл: %w", err)
		s.logger.Error(ctx, err)
		return err
	}

	// Successfully deleted
	return nil
}

func (s *S3Service) RedirectPath(bucket string, key string) string {
	return fmt.Sprintf(s.redirectFormat, bucket, key)
}

func (s *S3Service) FileName(id uuid.UUID) string {
	return s.FileNameS(id.String())
}

func (s *S3Service) FileNameS(id string) string {
	return fmt.Sprintf(s.fileFormat, id)
}
