package main

import (
	"context"
	"fmt"
	"github.com/budka-tech/configo"
	"github.com/budka-tech/envo"
	"github.com/budka-tech/logit-go"
	"s3n/internal/config"
	"s3n/internal/db"
	"s3n/internal/db/repository"
	"s3n/internal/endpoint"
	"s3n/internal/image_processing"
	"s3n/internal/s3"
)

func main() {
	const op = "main"
	cfg := configo.MustLoad[config.Config]()
	env, err := envo.New(cfg.Env)
	logger := logit.MustNewLogger(&cfg.App, &cfg.Logger, &cfg.Sentry, env)

	if err != nil {
		logger.Fatal(context.Background(), err)
		panic(err)
	}
	ctx := logger.NewCtx(context.Background(), op, nil)
	logger.Info(ctx, "Приложение успешно запущено")

	pool, err := db.NewClient(ctx, &cfg.DB)
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("ошибка при подключении БД: %s", err))
		panic(err)
	}
	logger.Info(ctx, "БД сервис успешно запущен")

	dbService := db.NewDBService(repository.NewPostgresRepository(pool))
	_ = dbService

	s3Service, err := s3.NewS3Service(ctx, logger, &cfg.S3Service)
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("ошибка при подключении S3: %s", err))
		panic(err)
	}
	logger.Info(ctx, "S3 сервис успешно запущен")
	_ = s3Service

	imageService := image_processing.NewImageService(&cfg.ImageProcessing, logger)
	_ = imageService

	endpointService, err := endpoint.NewEndpoint(ctx, s3Service, dbService, imageService, logger)
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("ошибка при создании эндпойнта: %s", err))
		panic(err)
	}

	grpcServer := endpoint.NewGrpcServer(endpointService, logger)
	logger.Info(ctx, "grpc сервер успешно запущен")

	go func() {
		err := grpcServer.Run(ctx)
		if err != nil {
			logger.Fatal(ctx, fmt.Errorf("ошибка при работе grpc: %s", err))
			panic(err)
		}
	}()

	server := endpoint.NewRedirectServer(s3Service, &cfg.HttpRedirect, logger)

	logger.Info(ctx, "redirect сервер запущен")
	err = server.Run(ctx)
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("ошибка при работе redirect сервера: %s", err))
		panic(err)
	}
}
