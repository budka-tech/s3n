package endpoint

import (
	"S3Proxy/internal/config"
	"S3Proxy/internal/s3"
	"context"
	"fmt"
	"github.com/budka-tech/logit-go"
	chi "github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type RedirectServer struct {
	router    *chi.Mux
	s3Service *s3.S3Service
	logger    logit.Logger
	port      int
}

func NewRedirectServer(s3Service *s3.S3Service, config *config.HttpRedirectConfig, logger logit.Logger) *RedirectServer {
	r := chi.NewRouter()

	s := &RedirectServer{
		router:    r,
		s3Service: s3Service,
		logger:    logger,
		port:      config.Port,
	}
	r.HandleFunc(config.PathPrefix+"/{bucket}/{filename}", s.redirectHandler)

	return s
}

func (s *RedirectServer) Run(ctx context.Context) error {
	const op = "RedirectServer.Run"
	ctx = s.logger.NewOpCtx(ctx, op)

	err := http.ListenAndServe(":"+strconv.Itoa(s.port), s.router)
	if err != nil {
		err = fmt.Errorf("ошибка при работе http сервера: %w", err)
		s.logger.Error(ctx, err)
		return err
	}

	return nil
}

func (s *RedirectServer) redirectHandler(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	filename := chi.URLParam(r, "filename")

	// Perform the redirect
	http.Redirect(w, r, s.s3Service.RedirectPath(bucket, s.s3Service.FileNameS(filename)), http.StatusFound) // StatusFound (302) for temporary redirects
}
