package config

import (
	"github.com/budka-tech/configo"
)

type Config struct {
	Env             string                `yaml:"env" env-default:"local"`
	App             configo.App           `yaml:"app"`
	Logger          configo.Logger        `yaml:"logger"`
	Sentry          configo.Sentry        `yaml:"sentry"`
	DB              configo.Database      `yaml:"database"`
	S3Service       S3ServiceConfig       `yaml:"s3"`
	ImageProcessing ImageProcessingConfig `yaml:"imageProcessing"`
	HttpRedirect    HttpRedirectConfig    `yaml:"httpRedirect"`
}

type S3ServiceConfig struct {
	S3Server configo.S3 `yaml:"s3Server"`

	// 1 аргумент - бакет
	// 2 аргумент - файл
	RedirectFormat string `yaml:"redirectFormat" env-required:"true"`
	FileFormat     string `yaml:"fileFormat" env-required:"true"`

	// менять только если нужно
	UploadGoroutines int   `yaml:"uploadGoroutines" env-default:"0"`
	UploadPartSize   int64 `yaml:"uploadPartSize" env-default:"0"`
}

type ImageProcessingConfig struct {
	DefaultQuality float32 `yaml:"defaultQuality" env-required:"true"`
	DefaultMaxSize int     `yaml:"defaultMaxSize" env-required:"true"`
}

type HttpRedirectConfig struct {
	Port       int    `yaml:"port" env-required:"true"`
	PathPrefix string `yaml:"pathPrefix" env-default:""`
}
