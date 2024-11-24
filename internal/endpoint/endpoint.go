package endpoint

import (
	"context"
	"fmt"
	"github.com/budka-tech/logit-go"
	"github.com/budka-tech/snip-common-go/status"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"math"
	"s3n/internal/db"
	"s3n/internal/db/models"
	"s3n/internal/endpoint/api_models"
	"s3n/internal/image_processing"
	"s3n/internal/s3"
	"sync"
)

type Endpoint struct {
	s3Service       *s3.S3Service
	dbService       *db.DBService
	imageService    *image_processing.ImageService
	logger          logit.Logger
	bucketCache     map[string]int16
	bucketCacheLock sync.RWMutex
}

func imageToAPI(image *models.Image) *api_models.Image {
	if image == nil {
		return nil
	}

	return &api_models.Image{
		ID: image.ID,
	}
}

func imageWithBucketToAPI(image *models.Image, bucketName string) *api_models.ImageWithBucket {
	if image == nil {
		return nil
	}

	return &api_models.ImageWithBucket{
		ID:         image.ID,
		BucketName: bucketName,
	}
}

func bucketToAPI(bucket *models.Bucket) *api_models.Bucket {
	if bucket == nil {
		return nil
	}

	return &api_models.Bucket{
		BucketName: bucket.BucketName,
	}
}

func NewEndpoint(ctx context.Context, s3Service *s3.S3Service, dbService *db.DBService, imageService *image_processing.ImageService, logger logit.Logger) (*Endpoint, error) {
	const op = "Endpoint.NewEndpoint"
	ctx = logger.NewOpCtx(ctx, op)

	buckets, err := dbService.GetAllBuckets(ctx, math.MaxUint16)
	if err != nil {
		err = fmt.Errorf("не удалось получить бакеты в БД: %w", err)
		logger.Error(ctx, err)
		return nil, err
	}
	var bucketCache map[string]int16
	for _, bucket := range buckets {
		bucketCache[bucket.BucketName] = bucket.ID
	}

	return &Endpoint{
		s3Service:    s3Service,
		dbService:    dbService,
		imageService: imageService,
		logger:       logger,
		bucketCache:  bucketCache,
	}, nil
}

func (e *Endpoint) RegisterBucket(ctx context.Context, bucketName string) (*api_models.Bucket, status.Status) {
	const op = "Endpoint.RegisterBucket"
	ctx = e.logger.NewOpCtx(ctx, op)

	bucket, err := e.dbService.CreateBucket(ctx, bucketName)
	if err != nil {
		err = fmt.Errorf("не удалось добавить бакет в БД: %w", err)
		e.logger.Error(ctx, err, zap.String("bucket_name", bucketName))
		return nil, status.InternalError
	}

	e.bucketCacheLock.Lock()
	e.bucketCache[bucket.BucketName] = bucket.ID
	e.bucketCacheLock.Unlock()

	return bucketToAPI(bucket), status.OK
}

func (e *Endpoint) HasBucket(ctx context.Context, bucketName string) (bool, status.Status) {
	const op = "Endpoint.GetBucket"
	ctx = e.logger.NewOpCtx(ctx, op)

	e.bucketCacheLock.RLock()
	_, ok := e.bucketCache[bucketName]
	e.bucketCacheLock.RUnlock()

	return ok, status.OK
}

func (e *Endpoint) UnregisterBucket(ctx context.Context, bucketName string) status.Status {
	const op = "Endpoint.UnregisterBucket"
	ctx = e.logger.NewOpCtx(ctx, op)

	e.bucketCacheLock.Lock()
	{
		id, ok := e.bucketCache[bucketName]
		if !ok {
			err := fmt.Errorf("не удалось найти бакет в кеше")
			e.logger.Error(ctx, err, zap.String("bucket_name", bucketName))
			return status.NotFound
		}

		err := e.dbService.DeleteBucket(ctx, id)
		if err != nil {
			err = fmt.Errorf("не удалось удалить бакет из БД: %w", err)
			e.logger.Error(ctx, err, zap.String("bucket_name", bucketName))
			return status.InternalError
		}

		delete(e.bucketCache, bucketName)
	}
	e.bucketCacheLock.Unlock()

	return status.OK
}

func (e *Endpoint) GetAllBuckets(ctx context.Context) ([]api_models.Bucket, status.Status) {
	const op = "Endpoint.GetAllBuckets"
	ctx = e.logger.NewOpCtx(ctx, op)

	var buckets []api_models.Bucket
	e.bucketCacheLock.RLock()
	{
		for name, _ := range e.bucketCache {
			buckets = append(buckets, api_models.Bucket{
				BucketName: name,
			})
		}
	}
	e.bucketCacheLock.RUnlock()

	return buckets, status.NotFound
}

func (e *Endpoint) CreateImage(ctx context.Context, bucketName string, file []byte, fileExtension string, quality *float32, maxSize *int) (*api_models.Image, status.Status) {
	const op = "Endpoint.CreateImage"
	ctx = e.logger.NewOpCtx(ctx, op)

	e.bucketCacheLock.RLock()
	id, ok := e.bucketCache[bucketName]
	e.bucketCacheLock.RUnlock()
	if !ok {
		err := fmt.Errorf("не удалось найти бакет в кеше")
		e.logger.Error(ctx, err, zap.String("bucket_name", bucketName))
		return nil, status.NotFound
	}

	processedFile, err := e.imageService.Transform(ctx, file, fileExtension, quality, maxSize)
	if err != nil {
		err = fmt.Errorf("не удалось обработать изображение: %w", err)
		e.logger.Error(ctx, err, zap.String("bucket_name", bucketName))
		return nil, status.InternalError
	}

	image, err := e.dbService.CreateImage(ctx, id)
	if err != nil {
		err = fmt.Errorf("не удалось добавить изображение в БД: %w", err)
		e.logger.Error(ctx, err, zap.String("bucket_name", bucketName))
		return nil, status.InternalError
	}

	err = e.s3Service.UploadFileBytes(ctx, bucketName, e.s3Service.FileName(image.ID), processedFile)
	if err != nil {
		err = fmt.Errorf("не удалось загрузить файл на S3: %w", err)
		e.logger.Error(ctx, err, zap.String("bucket_name", bucketName), zap.String("image_id", image.ID.String()))
		{
			err := e.dbService.DeleteBucket(ctx, id)
			if err != nil {
				err = fmt.Errorf("не удалось очистить изображение в БД: %w", err)
				e.logger.Error(ctx, err, zap.String("bucket_name", bucketName), zap.String("image_id", image.ID.String()))
				return nil, status.InternalError
			}
		}
		return nil, status.InternalError
	}

	return imageToAPI(image), status.OK
}

func (e *Endpoint) GetImage(ctx context.Context, id uuid.UUID) (*api_models.Image, status.Status) {
	const op = "Endpoint.GetImage"
	ctx = e.logger.NewOpCtx(ctx, op)

	image, err := e.dbService.GetImage(ctx, id)
	if err != nil {
		err = fmt.Errorf("не удалось получить изображение из БД: %w", err)
		e.logger.Error(ctx, err, zap.String("image_id", id.String()))
		return nil, status.NotFound
	}

	return imageToAPI(image), status.OK
}

func (e *Endpoint) GetImageWithBucket(ctx context.Context, id uuid.UUID) (*api_models.ImageWithBucket, status.Status) {
	const op = "Endpoint.GetImage"
	ctx = e.logger.NewOpCtx(ctx, op)

	image, bucket, err := e.dbService.GetImageWithBucket(ctx, id)
	if err != nil {
		err = fmt.Errorf("не удалось получить изображение из БД: %w", err)
		e.logger.Error(ctx, err, zap.String("image_id", id.String()))
		return nil, status.NotFound
	}

	return imageWithBucketToAPI(image, bucket.BucketName), status.OK
}

func (e *Endpoint) DeleteImage(ctx context.Context, id uuid.UUID) status.Status {
	const op = "Endpoint.DeleteImage"
	ctx = e.logger.NewOpCtx(ctx, op)

	image, bucket, err := e.dbService.GetImageWithBucket(ctx, id)
	if err != nil {
		err = fmt.Errorf("не удалось получить изображение из БД: %w", err)
		e.logger.Error(ctx, err, zap.String("image_id", id.String()))
		return status.NotFound
	}

	err = e.s3Service.DeleteFile(ctx, bucket.BucketName, e.s3Service.FileName(image.ID))
	if err != nil {
		err = fmt.Errorf("не удалось удалить изображение из БД: %w", err)
		e.logger.Error(ctx, err, zap.String("bucket_name", bucket.BucketName), zap.String("image_id", id.String()))
		return status.InternalError
	}

	err = e.dbService.DeleteImage(ctx, id)
	if err != nil {
		err = fmt.Errorf("не удалось удалить изображене из БД: %w", err)
		e.logger.Error(ctx, err, zap.String("bucket_name", bucket.BucketName), zap.String("image_id", id.String()))
		return status.InternalError
	}

	return status.OK
}

func (e *Endpoint) GetAllImages(ctx context.Context, limit int) ([]api_models.Image, status.Status) {
	const op = "Endpoint.GetAllImages"
	ctx = e.logger.NewOpCtx(ctx, op)

	images, err := e.dbService.GetAllImages(ctx, limit)
	if err != nil {
		err = fmt.Errorf("не удалось получить изображения зи БД: %w", err)
		e.logger.Error(ctx, err)
		return nil, status.InternalError
	}

	var apiImages []api_models.Image
	for _, image := range images {
		apiImages = append(apiImages, *imageToAPI(&image))
	}

	return apiImages, status.OK
}

func (e *Endpoint) GetImagesInBucket(ctx context.Context, bucketName string, limit int) ([]api_models.Image, status.Status) {
	const op = "Endpoint.GetImagesInBucket"
	ctx = e.logger.NewOpCtx(ctx, op)

	e.bucketCacheLock.RLock()
	id, ok := e.bucketCache[bucketName]
	e.bucketCacheLock.RUnlock()
	if !ok {
		err := fmt.Errorf("не удалось найти бакет в кеше")
		e.logger.Error(ctx, err, zap.String("bucket_name", bucketName))
		return nil, status.NotFound
	}

	images, err := e.dbService.GetImagesByBucketID(ctx, id, limit)
	if err != nil {
		err := fmt.Errorf("не удалось получить изображения: %w", err)
		e.logger.Error(ctx, err, zap.String("bucket_name", bucketName))
		return nil, status.NotFound
	}

	var apiImages []api_models.Image
	for _, image := range images {
		apiImages = append(apiImages, *imageToAPI(&image))
	}

	return apiImages, status.OK
}
