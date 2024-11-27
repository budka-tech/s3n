package image_processing

import (
	"bytes"
	"context"
	"fmt"
	"github.com/budka-tech/logit-go"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/nfnt/resize"
	"go.uber.org/zap"
	"image"
	"image/jpeg"
	"image/png"
	"s3n/internal/config"
)

type ImageService struct {
	DefaultQuality float32
	DefaultMaxSize int

	logger logit.Logger
}

func NewImageService(config *config.ImageProcessingConfig, logger logit.Logger) Service {
	return &ImageService{
		DefaultQuality: config.DefaultQuality,
		DefaultMaxSize: config.DefaultMaxSize,
		logger:         logger,
	}
}

func (s *ImageService) Transform(ctx context.Context, file []byte, fileFormat string, quality *float32, maxSize *int) ([]byte, error) {
	const op = "ImageService.Transform"
	ctx = s.logger.NewOpCtx(ctx, op)

	var resQuality float32
	if quality != nil {
		resQuality = *quality
	} else {
		resQuality = s.DefaultQuality
	}
	var resMaxSize int
	if maxSize != nil {
		resMaxSize = *maxSize
	} else {
		resMaxSize = s.DefaultMaxSize
	}

	var img image.Image
	var err error
	switch fileFormat {
	case "png":
		img, err = png.Decode(bytes.NewReader(file))
	case "jpeg", "jpg":
		img, err = jpeg.Decode(bytes.NewReader(file))
	case "webp":
		img, err = webp.Decode(bytes.NewReader(file), &decoder.Options{})
	default:
		err = fmt.Errorf("неизвестный формат изображения")
	}
	if err != nil {
		err = fmt.Errorf("не удалось прочитать изображение: %w", err)
		s.logger.Error(ctx, err, zap.String("format", fileFormat), zap.Int("file_size", len(file)))
		return nil, err
	}

	if img.Bounds().Size().X > resMaxSize || img.Bounds().Size().Y > resMaxSize {
		var newXSize int
		var newYSize int

		if img.Bounds().Size().X > img.Bounds().Size().Y {
			ratio := float32(img.Bounds().Size().Y) / float32(img.Bounds().Size().X)

			newXSize = resMaxSize
			newYSize = int(float32(newXSize) * ratio)

			if newYSize > resMaxSize {
				newYSize = resMaxSize
			}
		} else {
			ratio := float32(img.Bounds().Size().X) / float32(img.Bounds().Size().Y)

			newYSize = resMaxSize
			newXSize = int(float32(newYSize) * ratio)

			if newXSize > resMaxSize {
				newXSize = resMaxSize
			}
		}

		img = resize.Resize(uint(newXSize), uint(newYSize), img, resize.Lanczos3)
	}

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, resQuality)
	if err != nil {
		err = fmt.Errorf("не удалось создать webp encoder: %w", err)
		s.logger.Error(ctx, err)
		return nil, err
	}

	outBytes := &bytes.Buffer{}

	if err := webp.Encode(outBytes, img, options); err != nil {
		err = fmt.Errorf("не удался экспорт webp: %w", err)
		s.logger.Error(ctx, err)
		return nil, err
	}

	return outBytes.Bytes(), nil
}
