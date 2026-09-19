package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTP              HTTPConfig
	GRPC              GRPCConfig
	Token             TokenConfig
	InternalAPIKey    string
	PostgresDSN       string
	FirebaseProjectID string
}

type HTTPConfig struct {
	Addr string
}

type GRPCConfig struct {
	Addr string
}

type TokenConfig struct {
	Issuer        string
	ActiveKeyID   string
	PrivateKeyB64 string
	VerifyKeys    map[string]string
	TTL           time.Duration
}

func Load() (Config, error) {
	ttl, err := time.ParseDuration(envOrDefault("ENV_TOKEN_TTL", "5m"))
	if err != nil || ttl <= 0 {
		return Config{}, fmt.Errorf("ENV_TOKEN_TTL must be a positive duration")
	}
	verifyKeys := map[string]string{}
	if value := os.Getenv("ENV_TOKEN_VERIFY_KEYS_JSON"); value != "" {
		if err := json.Unmarshal([]byte(value), &verifyKeys); err != nil {
			return Config{}, fmt.Errorf("parse ENV_TOKEN_VERIFY_KEYS_JSON: %w", err)
		}
	}
	cfg := Config{
		HTTP: HTTPConfig{Addr: envOrDefault("ENV_HTTP_ADDR", ":8080")},
		GRPC: GRPCConfig{Addr: envOrDefault("ENV_GRPC_ADDR", ":9090")},
		Token: TokenConfig{
			Issuer:        envOrDefault("ENV_TOKEN_ISSUER", "beaesthetic-identity"),
			ActiveKeyID:   os.Getenv("ENV_TOKEN_ACTIVE_KEY_ID"),
			PrivateKeyB64: os.Getenv("ENV_TOKEN_PRIVATE_KEY_B64"),
			VerifyKeys:    verifyKeys,
			TTL:           ttl,
		},
		InternalAPIKey:    os.Getenv("ENV_INTERNAL_API_KEY"),
		PostgresDSN:       os.Getenv("ENV_POSTGRES_DSN"),
		FirebaseProjectID: os.Getenv("ENV_FIREBASE_PROJECT_ID"),
	}
	if cfg.Token.ActiveKeyID == "" || cfg.Token.PrivateKeyB64 == "" {
		return Config{}, fmt.Errorf("ENV_TOKEN_ACTIVE_KEY_ID and ENV_TOKEN_PRIVATE_KEY_B64 are required")
	}
	if cfg.InternalAPIKey == "" {
		return Config{}, fmt.Errorf("ENV_INTERNAL_API_KEY is required")
	}
	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("ENV_POSTGRES_DSN is required")
	}
	if cfg.FirebaseProjectID == "" {
		return Config{}, fmt.Errorf("ENV_FIREBASE_PROJECT_ID is required")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
