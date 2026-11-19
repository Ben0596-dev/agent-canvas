package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	Environment string
	AllowNoDB   bool
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr:    value("HTTP_ADDR", ":8080"),
		DatabaseURL: value("DATABASE_URL", ""),
		Environment: value("APP_ENV", "development"),
		AllowNoDB:   value("ALLOW_NO_DB", "true") == "true",
	}
	if c.DatabaseURL == "" && !c.AllowNoDB {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return c, nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
