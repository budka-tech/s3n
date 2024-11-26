package main

import (
	"context"
	pb "github.com/budka-tech/snip-common-go/contract/s3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("использование:\nexe <grpc адрес> <бакет>\n")
	}
	address := os.Args[1]
	bucket := os.Args[2]

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewEndpointClient(conn)

	inBucket, err := client.GetImagesInBucket(context.Background(), &pb.GetImagesInBucketRequest{
		BucketName: bucket,
		Limit:      math.MaxInt32,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("всего %d\n", len(inBucket.Images))

	for _, image := range inBucket.Images {
		if image != nil {
			_, err := client.DeleteImage(context.Background(), &pb.DeleteImageRequest{
				Id: image.Id,
			})
			if err != nil {
				log.Printf("ошибка удаления %s\n", image.Id)
			}
		}
	}
	log.Println("завершено")
}
