package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	MigrationsPath     string
	AllowedOrigins     []string
	RateLimitPerMinute int
	Environment        string
	WebAuthnRPID       string
	WebAuthnRPOrigins  []string
	VAPIDPublicKey     string
	VAPIDPrivateKey    string
}

func Load() *Config {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://moneyvault:moneyvault@localhost:5432/moneyvault?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		AccessTokenExpiry:  getDuration("ACCESS_TOKEN_EXPIRY", 15*time.Minute),
		RefreshTokenExpiry: getDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		MigrationsPath:     getEnv("MIGRATIONS_PATH", "file://migrations"),
		AllowedOrigins:     []string{getEnv("ALLOWED_ORIGIN", "http://localhost:5173")},
		RateLimitPerMinute: getInt("RATE_LIMIT_PER_MINUTE", 100),
		Environment:        getEnv("ENVIRONMENT", "development"),
		WebAuthnRPID:       getEnv("WEBAUTHN_RP_ID", "localhost"),
		WebAuthnRPOrigins:  []string{getEnv("WEBAUTHN_RP_ORIGIN", getEnv("ALLOWED_ORIGIN", "http://localhost:5173"))},
		VAPIDPublicKey:     getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey:    getEnv("VAPID_PRIVATE_KEY", ""),
	}

	if cfg.Environment == "production" {
		if cfg.JWTSecret == "change-me-in-production" || len(cfg.JWTSecret) < 32 {
			log.Fatal(fmt.Sprintf(
				"FATAL: JWT_SECRET is insecure (length=%d). In production, JWT_SECRET must be at least 32 characters and not the default value.",
				len(cfg.JWTSecret),
			))
		}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		d, err := time.ParseDuration(val)
		if err == nil {
			return d
		}
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		i, err := strconv.Atoi(val)
		if err == nil {
			return i
		}
	}
	return fallback
}
