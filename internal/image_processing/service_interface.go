package image_processing

import "context"

type Service interface {
	Transform(ctx context.Context, file []byte, fileFormat string, quality *float32, maxSize *int) ([]byte, error)
}
