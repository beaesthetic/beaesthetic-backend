package config

import (
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `koanf:"server"`
	MongoDB  MongoDBConfig  `koanf:"mongodb"`
	Consent  ConsentConfig  `koanf:"consent"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port    int    `koanf:"port"`
	BaseURL string `koanf:"base_url"`
}

// MongoDBConfig represents MongoDB configuration
type MongoDBConfig struct {
	ConnectionString string `koanf:"connection_string"`
	Database         string `koanf:"database"`
}

// ConsentConfig represents consent-specific configuration
type ConsentConfig struct {
	DefaultLinkExpiryHours int `koanf:"default_link_expiry_hours"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	// Load from YAML file if path is provided
	if configPath != "" {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			// Config file is optional, don't fail if not found
			_ = err
		}
	}

	// Load from environment variables (overrides file config)
	// Environment variable format: CONSENT_SERVER_PORT, CONSENT_MONGODB_CONNECTION_STRING, etc.
	if err := k.Load(env.Provider("CONSENT_", ".", func(s string) string {
		return strings.Replace(
			strings.ToLower(strings.TrimPrefix(s, "CONSENT_")),
			"_",
			".",
			-1,
		)
	}), nil); err != nil {
		return nil, err
	}

	var config Config

	// Set defaults
	config.Server.Port = 8080
	config.Server.BaseURL = "http://localhost:8080"
	config.MongoDB.ConnectionString = "mongodb://localhost:27017"
	config.MongoDB.Database = "consents"
	config.Consent.DefaultLinkExpiryHours = 48

	// Unmarshal into struct
	if err := k.Unmarshal("", &config); err != nil {
		return nil, err
	}

	return &config, nil
}
