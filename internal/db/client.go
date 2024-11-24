package db

import (
	"context"
	"fmt"
	"github.com/budka-tech/configo"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func NewClient(ctx context.Context, cfg *configo.Database) (pool *pgxpool.Pool, err error) {
	dsn := dsn(cfg)

	err = try(func() error {
		ctx, cancel := context.WithTimeout(ctx, cfg.AttemptDelay)
		defer cancel()

		pool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("Ошибка при подключении к базе данных")
		}

		return nil
	}, cfg.MaxAttempts, cfg.AttemptDelay)

	if err != nil {
		return nil, fmt.Errorf("Не удалось подключиться к базе данных после %v попыток\n", cfg.MaxAttempts)
	}

	return pool, nil
}

func dsn(cfg *configo.Database) string {
	return fmt.Sprintf("%v://%v:%v@%v:%v/%v", cfg.Type, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}

func try(fn func() error, attempts int, delay time.Duration) (err error) {
	for attempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attempts--

			continue
		}

		return nil
	}

	return
}
