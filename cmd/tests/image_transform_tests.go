package main

import (
	"context"
	"fmt"
	"github.com/budka-tech/logit-go"
	"io"
	"log"
	"os"
	"path/filepath"
	"s3n/internal/config"
	"s3n/internal/image_processing"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 5 {
		log.Fatal("usage:\nexe <путь к изображению> <качество> <максимальный размер> <количество циклов>\n")
	}

	filePath := os.Args[1]
	quality := os.Args[2]
	maxSize := os.Args[3]
	cycles := os.Args[4]

	fileExtension := filepath.Ext(filePath)[1:]

	qualityF64, err := strconv.ParseFloat(quality, 32)
	if err != nil {
		log.Fatal(err)
	}
	qualityF32 := float32(qualityF64)

	maxSizeI64, err := strconv.ParseInt(maxSize, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	maxSizeInt := int(maxSizeI64)

	cyclesI64, err := strconv.ParseInt(cycles, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	cyclesInt := int(cyclesI64)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return
	}

	imageService := image_processing.NewImageService(&config.ImageProcessingConfig{
		DefaultQuality: 0,
		DefaultMaxSize: 0,
	}, logit.NewNopLogger())

	ctx := context.Background()

	var maxTime time.Duration = 0
	var minTime time.Duration = time.Hour
	var totalTime time.Duration = 0

	fmt.Println("замер времени...")
	for i := 0; i < cyclesInt; i++ {
		start := time.Now()
		transform, err := imageService.Transform(ctx, bytes, fileExtension, &qualityF32, &maxSizeInt)
		elapsed := time.Since(start)
		if err != nil {
			panic(err)
		}
		_ = transform
		if elapsed > maxTime {
			maxTime = elapsed
		}
		if elapsed < minTime {
			minTime = elapsed
		}
		totalTime += elapsed
	}
	fmt.Printf("циклов: %d\n", cyclesInt)
	fmt.Printf("всего времени: %s\n", totalTime)
	fmt.Printf("среднее время: %s\n", totalTime/time.Duration(cyclesInt))
	fmt.Printf("максимум: %s\n", maxTime)
	fmt.Printf("минимум: %s\n", minTime)
}
