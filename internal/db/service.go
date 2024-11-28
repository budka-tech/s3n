package db

import (
	"context"
	"github.com/google/uuid"
	"s3n/internal/db/models"
	"s3n/internal/db/repository"
)

// DBService использует репозиторий для операций с bucket
type DBService struct {
	repo repository.Repository
}

// NewDBService создаёт новый DBService
func NewDBService(repo repository.Repository) Service {
	return &DBService{repo: repo}
}

// CreateBucket создает новый бакет
func (s *DBService) CreateBucket(ctx context.Context, bucketName string) (*models.Bucket, error) {
	return s.repo.InsertBucket(ctx, bucketName)
}

// GetBucket получает бакет по ID
func (s *DBService) GetBucket(ctx context.Context, id int16) (*models.Bucket, error) {
	return s.repo.GetBucketByID(ctx, id)
}

// DeleteBucket удаляет бакет по ID
func (s *DBService) DeleteBucket(ctx context.Context, id int16) error {
	return s.repo.DeleteBucketByID(ctx, id)
}

// GetAllBuckets получает все бакеты с лимитом на количество
func (s *DBService) GetAllBuckets(ctx context.Context, limit int) ([]models.Bucket, error) {
	return s.repo.GetAllBuckets(ctx, limit)
}

// CreateImage создает новое изображение в указанном бакете
func (s *DBService) CreateImage(ctx context.Context, bucketID int16) (*models.Image, error) {
	return s.repo.InsertImage(ctx, bucketID)
}

func (s *DBService) AddImage(ctx context.Context, bucketID int16, id uuid.UUID) error {
	return s.repo.AddImage(ctx, bucketID, id)
}

// GetImage получает изображение по ID
func (s *DBService) GetImage(ctx context.Context, id uuid.UUID) (*models.Image, error) {
	return s.repo.GetImageByID(ctx, id)
}

// GetImageWithBucket получает изображение по ID
func (s *DBService) GetImageWithBucket(ctx context.Context, id uuid.UUID) (*models.Image, *models.Bucket, error) {
	return s.repo.GetImageWithBucket(ctx, id)
}

// DeleteImage удаляет изображение по ID
func (s *DBService) DeleteImage(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteImageByID(ctx, id)
}

// GetAllImages получает все изображения с лимитом на количество
func (s *DBService) GetAllImages(ctx context.Context, limit int) ([]models.Image, error) {
	return s.repo.GetAllImages(ctx, limit)
}

// GetImagesByBucketID получает все изображения из конкретного бакета с лимитом на количество
func (s *DBService) GetImagesByBucketID(ctx context.Context, bucketID int16, limit int) ([]models.Image, error) {
	return s.repo.GetImagesByBucketID(ctx, bucketID, limit)
}
