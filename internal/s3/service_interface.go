package s3

import (
	"context"
	"github.com/google/uuid"
	"io"
)

type Service interface {
	UploadFile(ctx context.Context, bucket string, key string, file io.Reader) error
	UploadFileBytes(ctx context.Context, bucket string, key string, file []byte) error
	DeleteFile(ctx context.Context, bucket string, key string) error
	RedirectPath(bucket string, key string) string
	FileName(id uuid.UUID) string
	FileNameS(id string) string
}
