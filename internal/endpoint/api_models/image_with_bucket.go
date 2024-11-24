package api_models

import "github.com/google/uuid"

type ImageWithBucket struct {
	ID         uuid.UUID // Уникальный идентификатор изображения
	BucketName string
}
