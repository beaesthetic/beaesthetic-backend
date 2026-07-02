package config

import (
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const keyDelimiter = "."

type Config struct {
	App        AppConfig        `koanf:"app"`
	HTTP       HTTPConfig       `koanf:"http"`
	Postgres   PostgresConfig   `koanf:"postgres"`
	Mongo      MongoConfig      `koanf:"mongo"`
	RabbitMQ   RabbitMQConfig   `koanf:"rabbitmq"`
	SMSGateway SMSGatewayConfig `koanf:"sms_gateway"`
}

type AppConfig struct {
	Name string `koanf:"name"`
	Env  string `koanf:"env"`
}

type HTTPConfig struct {
	Addr string `koanf:"addr"`
}

type PostgresConfig struct {
	DSN string `koanf:"dsn"`
}

type MongoConfig struct {
	URI        string `koanf:"uri"`
	Database   string `koanf:"database"`
	Collection string `koanf:"collection"`
}

type RabbitMQConfig struct {
	URL                      string        `koanf:"url"`
	NotificationQueue        string        `koanf:"notification_queue"`
	NotificationConfirmQueue string        `koanf:"notification_confirm_queue"`
	RetryTTL                 time.Duration `koanf:"retry_ttl"`
}

type SMSGatewayConfig struct {
	URL        string `koanf:"url"`
	APIKey     string `koanf:"api_key"`
	FromNumber string `koanf:"from_number"`
	WebhookURL string `koanf:"webhook_url"`
}

func Load(envFile string) (Config, error) {
	k := koanf.New(keyDelimiter)
	if err := loadDefaults(k); err != nil {
		return Config{}, err
	}
	if strings.TrimSpace(envFile) != "" {
		if err := k.Load(file.Provider(envFile), dotenv.ParserEnv("", keyDelimiter, normalizeEnvKey)); err != nil {
			return Config{}, err
		}
	}
	if err := k.Load(env.Provider("", keyDelimiter, normalizeEnvKey), nil); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadDefaults(k *koanf.Koanf) error {
	defaults := map[string]interface{}{
		"app.name":                            "beaesthetic-notifications",
		"app.env":                             "dev",
		"http.addr":                           ":8080",
		"mongo.database":                      "notifications",
		"mongo.collection":                    "notifications",
		"rabbitmq.notification_queue":         "NotificationQueue",
		"rabbitmq.notification_confirm_queue": "NotificationConfirmQueue",
		"rabbitmq.retry_ttl":                  "1ms",
		"sms_gateway.from_number":             "123",
	}
	for key, value := range defaults {
		if err := k.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func normalizeEnvKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(key, "__", keyDelimiter))
}
