package endpoint

import (
	"context"
	"fmt"
	"github.com/budka-tech/logit-go"
	commonv1 "github.com/budka-tech/snip-common-go/contract/common"
	pb "github.com/budka-tech/snip-common-go/contract/s3"
	"github.com/budka-tech/snip-common-go/port"
	st "github.com/budka-tech/snip-common-go/status"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"net"
	"s3n/internal/endpoint/api_models"
)

type GrpcServer struct {
	pb.UnimplementedEndpointServer
	endpoint *Endpoint
	logger   logit.Logger
}

func NewGrpcServer(endpoint *Endpoint, logger logit.Logger) *GrpcServer {
	return &GrpcServer{
		endpoint: endpoint,
		logger:   logger,
	}
}

func (g GrpcServer) Run(ctx context.Context) error {
	const op = "GrpcServer.Run"
	ctx = g.logger.NewOpCtx(ctx, op)

	lis, err := net.Listen("tcp", port.FormatTCP(port.S3Grpc))
	if err != nil {
		g.logger.Fatal(ctx, fmt.Errorf("ошибка запуска tcp слушителя: %s", err))
		return err
	}
	g.logger.Info(ctx, fmt.Sprintf("сервер запущен на: %v", lis.Addr()))

	grpcServer := grpc.NewServer()
	pb.RegisterEndpointServer(grpcServer, &g)
	if err := grpcServer.Serve(lis); err != nil {
		g.logger.Fatal(ctx, fmt.Errorf("ошибка grpc: %s", err))
		return err
	}
	return nil
}

func bucketToProto(bucket *api_models.Bucket) *pb.Bucket {
	if bucket == nil {
		return nil
	}
	return &pb.Bucket{
		BucketName: bucket.BucketName,
	}
}

func bucketsToProto(bucket []api_models.Bucket) []*pb.Bucket {
	var buckets []*pb.Bucket
	for _, bucket := range bucket {
		buckets = append(buckets, bucketToProto(&bucket))
	}
	return buckets
}

func imageToProto(image *api_models.Image) *pb.Image {
	if image == nil {
		return nil
	}
	return &pb.Image{
		Id: image.ID[:],
	}
}

func imagesToProto(image []api_models.Image) []*pb.Image {
	var images []*pb.Image
	for _, img := range image {
		images = append(images, imageToProto(&img))
	}
	return images
}

func imageWithBucketToProto(imageWithBucket *api_models.ImageWithBucket) *pb.ImageWithBucket {
	if imageWithBucket == nil {
		return nil
	}
	return &pb.ImageWithBucket{
		Id:         imageWithBucket.ID[:],
		BucketName: imageWithBucket.BucketName,
	}
}

func (g GrpcServer) RegisterBucket(ctx context.Context, request *pb.RegisterBucketRequest) (*pb.RegisterBucketResponse, error) {
	bucket, status := g.endpoint.RegisterBucket(ctx, request.BucketName)
	return &pb.RegisterBucketResponse{
		Bucket: bucketToProto(bucket),
		Status: status,
	}, nil
}

func (g GrpcServer) HasBucket(ctx context.Context, request *pb.HasBucketRequest) (*pb.HasBucketResponse, error) {
	exist, status := g.endpoint.HasBucket(ctx, request.BucketName)
	return &pb.HasBucketResponse{
		Exists: exist,
		Status: status,
	}, nil
}

func (g GrpcServer) UnregisterBucket(ctx context.Context, request *pb.UnregisterBucketRequest) (*commonv1.Response, error) {
	status := g.endpoint.UnregisterBucket(ctx, request.BucketName)
	return &commonv1.Response{
		Status: status,
	}, nil
}

func (g GrpcServer) GetAllBuckets(ctx context.Context, request *pb.GetAllBucketsRequest) (*pb.GetAllBucketsResponse, error) {
	buckets, status := g.endpoint.GetAllBuckets(ctx)
	return &pb.GetAllBucketsResponse{
		Buckets: bucketsToProto(buckets),
		Status:  status,
	}, nil
}

func (g GrpcServer) CreateImage(ctx context.Context, request *pb.CreateImageRequest) (*pb.CreateImageResponse, error) {
	var MaxSize *int
	if request.MaxSize != nil {
		maxSize := int(*request.MaxSize)
		MaxSize = &maxSize
	}
	img, status := g.endpoint.CreateImage(ctx, request.BucketName, request.File, request.FileExtension, request.Quality, MaxSize)
	return &pb.CreateImageResponse{
		Image:  imageToProto(img),
		Status: status,
	}, nil
}

func (g GrpcServer) GetImage(ctx context.Context, request *pb.GetImageRequest) (*pb.GetImageResponse, error) {
	const op = "GrpcServer.GetImage"
	ctx = g.logger.NewOpCtx(ctx, op)

	Id, err := uuid.FromBytes(request.Id)
	if err != nil {
		g.logger.Error(ctx, fmt.Errorf("ошибка парсинга uuid: %s", err))
		return &pb.GetImageResponse{
			Status: st.IncorrectValue,
		}, nil
	}
	img, status := g.endpoint.GetImage(ctx, Id)
	return &pb.GetImageResponse{
		Image:  imageToProto(img),
		Status: status,
	}, nil
}

func (g GrpcServer) GetImageWithBucket(ctx context.Context, request *pb.GetImageWithBucketRequest) (*pb.GetImageWithBucketResponse, error) {
	const op = "GrpcServer.GetImageWithBucket"
	ctx = g.logger.NewOpCtx(ctx, op)

	Id, err := uuid.FromBytes(request.Id)
	if err != nil {
		g.logger.Error(ctx, fmt.Errorf("ошибка парсинга uuid: %s", err))
		return &pb.GetImageWithBucketResponse{
			Status: st.IncorrectValue,
		}, nil
	}

	imageWithBucket, status := g.endpoint.GetImageWithBucket(ctx, Id)

	return &pb.GetImageWithBucketResponse{
		ImageWithBucket: imageWithBucketToProto(imageWithBucket),
		Status:          status,
	}, nil
}

func (g GrpcServer) DeleteImage(ctx context.Context, request *pb.DeleteImageRequest) (*commonv1.Response, error) {
	const op = "GrpcServer.DeleteImage"
	ctx = g.logger.NewOpCtx(ctx, op)

	Id, err := uuid.FromBytes(request.Id)
	if err != nil {
		g.logger.Error(ctx, fmt.Errorf("ошибка парсинга uuid: %s", err))
		return &commonv1.Response{
			Status: st.IncorrectValue,
		}, nil
	}

	status := g.endpoint.DeleteImage(ctx, Id)

	return &commonv1.Response{
		Status: status,
	}, nil
}

func (g GrpcServer) GetAllImages(ctx context.Context, request *pb.GetAllImagesRequest) (*pb.GetAllImagesResponse, error) {
	images, status := g.endpoint.GetAllImages(ctx, int(request.Limit))
	return &pb.GetAllImagesResponse{
		Images: imagesToProto(images),
		Status: status,
	}, nil
}

func (g GrpcServer) GetImagesInBucket(ctx context.Context, request *pb.GetImagesInBucketRequest) (*pb.GetImagesInBucketResponse, error) {
	images, status := g.endpoint.GetImagesInBucket(ctx, request.BucketName, int(request.Limit))
	return &pb.GetImagesInBucketResponse{
		Images: imagesToProto(images),
		Status: status,
	}, nil
}
