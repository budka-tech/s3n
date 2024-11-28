package db

import (
	"context"
	"github.com/google/uuid"
	"s3n/internal/db/models"
)

type Service interface {
	CreateBucket(ctx context.Context, bucketName string) (*models.Bucket, error)
	GetBucket(ctx context.Context, id int16) (*models.Bucket, error)
	DeleteBucket(ctx context.Context, id int16) error
	GetAllBuckets(ctx context.Context, limit int) ([]models.Bucket, error)
	CreateImage(ctx context.Context, bucketID int16) (*models.Image, error)
	AddImage(ctx context.Context, bucketID int16, id uuid.UUID) error
	GetImage(ctx context.Context, id uuid.UUID) (*models.Image, error)
	GetImageWithBucket(ctx context.Context, id uuid.UUID) (*models.Image, *models.Bucket, error)
	DeleteImage(ctx context.Context, id uuid.UUID) error
	GetAllImages(ctx context.Context, limit int) ([]models.Image, error)
	GetImagesByBucketID(ctx context.Context, bucketID int16, limit int) ([]models.Image, error)
}
