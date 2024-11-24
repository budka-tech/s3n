package main

import (
	"github.com/budka-tech/configo"
	"github.com/budka-tech/spg"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"s3n/internal/config"
)

func main() {
	cfg := configo.MustLoad[config.Config]()
	spg.MigrateCreate(cfg.DB)
}
