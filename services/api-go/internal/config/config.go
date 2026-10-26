// Package config loads and validates process configuration from the environment.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const PlaceholderJWTSecret = "change-me-dev-only-secret-change-me"
const PlaceholderWebhookSecret = "change-me-dev-webhook-secret-change-me"

type Config struct {
	AppEnv      string // dev | beta | production
	HTTPAddr    string
	APIMode     string // mock | db
	JWTSecret   string
	MySQLDSN    string
	RedisAddr   string
	MockPayment bool
	Version     string

	// WebhookSecret authenticates PSP callbacks (HMAC-SHA256). Empty disables the webhook endpoint.
	WebhookSecret string

	// S3-compatible object storage (MinIO in DEV) for presigned uploads. S3Endpoint is the URL clients can reach.
	S3Endpoint, S3Bucket, S3Region, S3AccessKey, S3SecretKey string
}

func getenv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}

// Load reads configuration from environment variables with DEV-friendly defaults.
func Load() Config {
	return Config{
		AppEnv:      getenv("APP_ENV", "dev"),
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		APIMode:     getenv("API_MODE", "mock"),
		JWTSecret:   getenv("JWT_SECRET", PlaceholderJWTSecret),
		MySQLDSN:    getenv("MYSQL_DSN", ""),
		RedisAddr:   getenv("REDIS_ADDR", ""),
		MockPayment: getenv("MOCK_PAYMENT", "true") == "true",
		Version:     getenv("APP_VERSION", "0.1.0-sprint1"),

		WebhookSecret: getenv("WEBHOOK_SECRET", ""),
		S3Endpoint:    getenv("S3_ENDPOINT", ""),
		S3Bucket:      getenv("S3_BUCKET", "tep-media"),
		S3Region:      getenv("S3_REGION", "us-east-1"),
		S3AccessKey:   getenv("S3_ACCESS_KEY", ""),
		S3SecretKey:   getenv("S3_SECRET_KEY", ""),
	}
}

// Validate implements the startup rules in docs/architecture/environment-strategy.md.
func (c Config) Validate() error {
	var errs []error
	switch c.AppEnv {
	case "dev", "beta", "production":
	default:
		errs = append(errs, fmt.Errorf("APP_ENV must be dev|beta|production, got %q", c.AppEnv))
	}
	if c.APIMode == "db" && c.MySQLDSN == "" {
		errs = append(errs, errors.New("API_MODE=db requires MYSQL_DSN"))
	}
	if c.APIMode != "mock" && c.APIMode != "db" {
		errs = append(errs, fmt.Errorf("API_MODE must be mock|db, got %q", c.APIMode))
	}
	if len(c.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 characters"))
	}
	if c.AppEnv != "dev" && c.JWTSecret == PlaceholderJWTSecret {
		errs = append(errs, errors.New("JWT_SECRET is the default placeholder; set a real secret outside dev"))
	}
	if c.WebhookSecret != "" && len(c.WebhookSecret) < 32 {
		errs = append(errs, errors.New("WEBHOOK_SECRET must be at least 32 characters"))
	}
	if c.AppEnv != "dev" && c.WebhookSecret == PlaceholderWebhookSecret {
		errs = append(errs, errors.New("WEBHOOK_SECRET is the default placeholder; set a real secret outside dev"))
	}
	if (c.S3AccessKey == "") != (c.S3SecretKey == "") {
		errs = append(errs, errors.New("S3_ACCESS_KEY and S3_SECRET_KEY must be set together"))
	}
	if c.AppEnv == "production" {
		if !c.MockPayment && c.WebhookSecret == "" {
			errs = append(errs, errors.New("WEBHOOK_SECRET is required in production (payment confirmations arrive by webhook)"))
		}
		if c.MockPayment {
			errs = append(errs, errors.New("MOCK_PAYMENT must be false in production"))
		}
		if c.APIMode == "mock" {
			errs = append(errs, errors.New("API_MODE=mock is not allowed in production"))
		}
		if !strings.Contains(c.MySQLDSN, "tls=") {
			errs = append(errs, errors.New("MYSQL_DSN must enable TLS (tls=true) in production"))
		}
	}
	return errors.Join(errs...)
}
