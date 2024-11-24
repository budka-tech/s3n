package repository

import (
	"context"
	"github.com/google/uuid"
	"s3n/internal/db/models"
)

// Repository определяет интерфейс для работы с bucket и image
type Repository interface {
	// Методы для Bucket
	InsertBucket(ctx context.Context, bucketName string) (*models.Bucket, error)
	GetBucketByID(ctx context.Context, id int16) (*models.Bucket, error)
	DeleteBucketByID(ctx context.Context, id int16) error
	GetAllBuckets(ctx context.Context, limit int) ([]models.Bucket, error)

	// Методы для Image
	InsertImage(ctx context.Context, bucketID int16) (*models.Image, error)
	GetImageByID(ctx context.Context, id uuid.UUID) (*models.Image, error)
	GetImageWithBucket(ctx context.Context, id uuid.UUID) (*models.Image, *models.Bucket, error)
	DeleteImageByID(ctx context.Context, id uuid.UUID) error
	GetAllImages(ctx context.Context, limit int) ([]models.Image, error)
	GetImagesByBucketID(ctx context.Context, bucketID int16, limit int) ([]models.Image, error)
}
