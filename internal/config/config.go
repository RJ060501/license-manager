package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TenantID     string
	ClientID     string
	ClientSecret string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		TenantID:     os.Getenv("TENANT_ID"),
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
	}

	if cfg.TenantID == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return cfg, nil
}
