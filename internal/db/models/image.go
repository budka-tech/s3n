package models

import (
	"github.com/google/uuid"
)

type Image struct {
	ID       uuid.UUID // Уникальный идентификатор изображения
	BucketID int16     // Внешний ключ на bucket
}
