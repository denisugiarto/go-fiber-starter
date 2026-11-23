// Package config membaca konfigurasi dari environment variable. Rahasia tidak pernah ditulis di kode.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Env         string
	Addr        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
	CORSOrigins []string
}

func Load() (Config, error) {
	c := Config{
		Env:         get("APP_ENV", "development"),
		Addr:        ":" + get("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		CORSOrigins: strings.Split(get("CORS_ORIGINS", "http://localhost:3000"), ","),
	}
	ttl, err := time.ParseDuration(get("JWT_TTL", "24h"))
	if err != nil {
		return c, fmt.Errorf("JWT_TTL tidak valid: %w", err)
	}
	c.JWTTTL = ttl
	switch {
	case c.DatabaseURL == "":
		return c, errors.New("DATABASE_URL wajib diisi")
	case len(c.JWTSecret) < 32:
		return c, errors.New("JWT_SECRET wajib diisi, minimal 32 karakter")
	}
	return c, nil
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
