package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	Environment string
}

var appConfig Config

func Load() (*Config, error) {
	_ = godotenv.Load("../../env/.env")

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT is not set")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	appConfig = Config{
		DatabaseURL: databaseURL,
		Port:        port,
		Environment: env,
	}
	return &appConfig, nil
}

func Get() *Config {
	return &appConfig
}
