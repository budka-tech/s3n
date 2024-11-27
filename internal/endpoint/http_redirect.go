package endpoint

import (
	"context"
	"fmt"
	"github.com/budka-tech/logit-go"
	chi "github.com/go-chi/chi/v5"
	"net/http"
	"s3n/internal/config"
	"s3n/internal/s3"
	"strconv"
)

type RedirectServer struct {
	router    *chi.Mux
	s3Service s3.Service
	logger    logit.Logger
	port      int
}

func NewRedirectServer(s3Service s3.Service, config *config.HttpRedirectConfig, logger logit.Logger) *RedirectServer {
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
