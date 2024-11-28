package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"s3n/internal/db/models"
)

// PostgresRepository — реализация Repository для PostgreSQL с использованием pgxpool
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository создает новый экземпляр PostgresRepository с пулом подключений
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &PostgresRepository{pool: pool}
}

// InsertBucket добавляет новый bucket в базу данных и возвращает его
func (r *PostgresRepository) InsertBucket(ctx context.Context, bucketName string) (*models.Bucket, error) {
	query := `INSERT INTO bucket (bucket_name) VALUES ($1) RETURNING id`
	var id int16
	err := r.pool.QueryRow(ctx, query, bucketName).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &models.Bucket{ID: id, BucketName: bucketName}, nil
}

// GetBucketByID возвращает bucket по его ID
func (r *PostgresRepository) GetBucketByID(ctx context.Context, id int16) (*models.Bucket, error) {
	var bucket models.Bucket
	query := `SELECT id, bucket_name FROM bucket WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&bucket.ID, &bucket.BucketName)
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// DeleteBucketByID удаляет bucket по его ID
func (r *PostgresRepository) DeleteBucketByID(ctx context.Context, id int16) error {
	query := `DELETE FROM bucket WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// GetAllBuckets возвращает список всех бакетов с ограничением на количество
func (r *PostgresRepository) GetAllBuckets(ctx context.Context, limit int) ([]models.Bucket, error) {
	query := `SELECT id, bucket_name FROM bucket LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []models.Bucket
	for rows.Next() {
		var bucket models.Bucket
		if err := rows.Scan(&bucket.ID, &bucket.BucketName); err != nil {
			return nil, err
		}
		buckets = append(buckets, bucket)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return buckets, nil
}

// InsertImage добавляет новое изображение в базу данных и возвращает его
func (r *PostgresRepository) InsertImage(ctx context.Context, bucketID int16) (*models.Image, error) {
	query := `INSERT INTO image (bucket_id) VALUES ($1) RETURNING id`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, bucketID).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &models.Image{ID: id, BucketID: bucketID}, nil
}

func (r *PostgresRepository) AddImage(ctx context.Context, bucketId int16, id uuid.UUID) error {
	query := `INSERT INTO image (bucket_id, id) VALUES ($1, $2) RETURNING id`
	err := r.pool.QueryRow(ctx, query, bucketId, id).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

// GetImageByID возвращает изображение по его ID
func (r *PostgresRepository) GetImageByID(ctx context.Context, id uuid.UUID) (*models.Image, error) {
	var image models.Image
	query := `SELECT id, bucket_id FROM image WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&image.ID, &image.BucketID)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageWithBucket извлекает изображение с данными о бакете по ID изображения
func (r *PostgresRepository) GetImageWithBucket(ctx context.Context, id uuid.UUID) (*models.Image, *models.Bucket, error) {
	query := `
        SELECT 
            i.id, 
            i.bucket_id, 
            b.id, 
            b.bucket_name
        FROM image i
        JOIN bucket b ON i.bucket_id = b.id
        WHERE i.id = $1
    `
	var image models.Image
	var bucket models.Bucket
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&image.ID,
		&image.BucketID,
		&bucket.ID,
		&bucket.BucketName,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil, errors.New("image not found")
		}
		return nil, nil, err
	}
	return &image, &bucket, nil
}

// DeleteImageByID удаляет изображение по его ID
func (r *PostgresRepository) DeleteImageByID(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM image WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// GetAllImages возвращает список всех изображений с ограничением на количество
func (r *PostgresRepository) GetAllImages(ctx context.Context, limit int) ([]models.Image, error) {
	query := `SELECT id, bucket_id FROM image LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var image models.Image
		if err := rows.Scan(&image.ID, &image.BucketID); err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return images, nil
}

// GetImagesByBucketID возвращает список изображений для заданного bucketID с ограничением на количество
func (r *PostgresRepository) GetImagesByBucketID(ctx context.Context, bucketID int16, limit int) ([]models.Image, error) {
	query := `SELECT id, bucket_id FROM image WHERE bucket_id = $1 LIMIT $2`
	rows, err := r.pool.Query(ctx, query, bucketID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var image models.Image
		if err := rows.Scan(&image.ID, &image.BucketID); err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return images, nil
}
