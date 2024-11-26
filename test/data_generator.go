package main

import (
	pb "github.com/budka-tech/snip-common-go/contract/s3"
	"github.com/gogo/protobuf/proto"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) != 6 {
		log.Fatal("использование:\nexe <входной файл> <итоговый файл> <бакет> <качество> <максимальный размер>\n")
	}
	inputFile := os.Args[1]
	outputFile := os.Args[2]
	bucket := os.Args[3]
	quality := os.Args[4]
	maxSize := os.Args[5]

	fileExtension := filepath.Ext(inputFile)[1:]

	qualityF64, err := strconv.ParseFloat(quality, 32)
	if err != nil {
		log.Fatal(err)
	}
	qualityF32 := float32(qualityF64)

	maxSizeI64, err := strconv.ParseInt(maxSize, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	maxSizeI32 := int32(maxSizeI64)

	oFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer oFile.Close()

	iFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer iFile.Close()

	bytes, err := io.ReadAll(iFile)
	if err != nil {
		return
	}

	marshal, err := proto.Marshal(&pb.CreateImageRequest{
		BucketName:    bucket,
		File:          bytes,
		FileExtension: fileExtension,
		Quality:       &qualityF32,
		MaxSize:       &maxSizeI32,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("размер", len(marshal))

	_, err = oFile.Write(marshal)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("успешно завершено")
}
