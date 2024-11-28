package main

import (
	"context"
	"fmt"
	"github.com/budka-tech/logit-go"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
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
	if len(os.Args) != 6 {
		log.Fatal("usage:\nexe <путь к изображению> <качество> <максимальный размер> <количество циклов> <время между замерами cpu и mem>\n")
	}

	filePath := os.Args[1]
	quality := os.Args[2]
	maxSize := os.Args[3]
	cycles := os.Args[4]
	intervalsString := os.Args[5]

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

	interval, err := time.ParseDuration(intervalsString)
	if err != nil {
		log.Fatal(err)
	}

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

	ctx, cancel := context.WithCancel(context.Background())

	beforeCpu, err := cpu.Get()
	if err != nil {
		log.Fatal(err)
	}
	beforeMem, err := memory.Get()
	if err != nil {
		log.Fatal(err)
	}

	_ = beforeMem
	_ = beforeCpu

	var maxTime time.Duration = 0
	var minTime time.Duration = time.Hour
	var totalTime time.Duration = 0

	var maxMemUsed = beforeMem.Used

	var maxCpuUsage = beforeCpu.User

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				nowMem, err := memory.Get()
				if err != nil {
					log.Fatal(err)
				}
				if nowMem.Used > maxMemUsed {
					maxMemUsed = nowMem.Used
				}
				nowCpu, err := cpu.Get()
				if err != nil {
					log.Fatal(err)
				}
				if nowCpu.User > maxCpuUsage {
					maxCpuUsage = nowCpu.User
				}
			case <-ctx.Done():
				return
			}
		}
	}()

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
	cancel()

	fmt.Printf("качество:            %f\n", qualityF32)
	fmt.Printf("максимальный размер: %d\n", maxSizeInt)
	fmt.Printf("размер файла:        %d\n", len(bytes))
	println()
	fmt.Printf("циклов:        %d\n", cyclesInt)
	fmt.Printf("всего времени: %s\n", totalTime)
	fmt.Printf("среднее время: %s\n", totalTime/time.Duration(cyclesInt))
	fmt.Printf("максимум:      %s\n", maxTime)
	fmt.Printf("минимум:       %s\n", minTime)
	println()
	fmt.Printf("максимум памяти: %d\n", maxMemUsed-beforeMem.Used)
	fmt.Printf("                 %f%%\n", float64(maxMemUsed-beforeMem.Used)/float64(beforeMem.Total)*100)
	println()
	fmt.Printf("максимум cpu-core: %f%%\n", float64(maxCpuUsage)/float64(beforeCpu.Total)*float64(beforeCpu.CPUCount)*100)
}
