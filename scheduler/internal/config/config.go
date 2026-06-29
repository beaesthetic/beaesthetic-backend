package config

import (
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server    ServerConfig    `koanf:"server"`
	Scheduler SchedulerConfig `koanf:"scheduler"`
	Postgres  PostgresConfig  `koanf:"postgres"`
	RabbitMQ  RabbitMQConfig  `koanf:"rabbitmq"`
	Redis     RedisConfig     `koanf:"redis"`
	Otel      OtelConfig      `koanf:"otel"`
	Log       LogConfig       `koanf:"log"`
}

type ServerConfig struct {
	Port int `koanf:"port"`
}

type SchedulerConfig struct {
	Name            string        `koanf:"name"`
	PollingInterval time.Duration `koanf:"polling_interval"`
	PeekLeaseTTL    time.Duration `koanf:"peek_lease_ttl"`
	PeekBatchSize   int           `koanf:"peek_batch_size"`
}

type PostgresConfig struct {
	DSN string `koanf:"dsn"`
}

type RabbitMQConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Exchange string `koanf:"exchange"`
}

type RedisConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

type OtelConfig struct {
	Endpoint string `koanf:"endpoint"`
	Enabled  bool   `koanf:"enabled"`
}

type LogConfig struct {
	Level       string `koanf:"level"`
	Development bool   `koanf:"development"`
}

func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	if configPath != "" {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			_ = err
		}
	}

	if err := k.Load(env.Provider("", ".", transformEnv), nil); err != nil {
		return nil, err
	}

	cfg := Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Scheduler: SchedulerConfig{
			Name:            "reminders",
			PollingInterval: 60 * time.Second,
			PeekLeaseTTL:    30 * time.Second,
			PeekBatchSize:   50,
		},
		Postgres: PostgresConfig{
			DSN: "postgres://postgres:postgres@localhost:5432/scheduler?sslmode=disable",
		},
		RabbitMQ: RabbitMQConfig{
			Host:     "localhost",
			Port:     5672,
			Username: "guest",
			Password: "guest",
			Exchange: "",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		Otel: OtelConfig{
			Endpoint: "localhost:4317",
			Enabled:  true,
		},
		Log: LogConfig{
			Level:       "info",
			Development: false,
		},
	}

	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	applyDuration(k, "scheduler.polling_interval", &cfg.Scheduler.PollingInterval)
	applyDuration(k, "scheduler.peek_lease_ttl", &cfg.Scheduler.PeekLeaseTTL)

	if cfg.RabbitMQ.Username == "" {
		cfg.RabbitMQ.Username = "guest"
	}
	if cfg.RabbitMQ.Password == "" {
		cfg.RabbitMQ.Password = "guest"
	}

	return &cfg, nil
}

func applyDuration(k *koanf.Koanf, key string, target *time.Duration) {
	if !k.Exists(key) {
		return
	}
	if d := k.Duration(key); d > 0 {
		*target = d
	}
}

func transformEnv(name string) string {
	name = strings.ToUpper(name)
	mapping := map[string]string{
		"SERVER_PORT":                  "server.port",
		"SCHEDULER_NAME":               "scheduler.name",
		"SCHEDULER_POLLING_INTERVAL":   "scheduler.polling_interval",
		"SCHEDULER_PEEK_LEASE_TTL":     "scheduler.peek_lease_ttl",
		"SCHEDULER_PEEK_BATCH_SIZE":    "scheduler.peek_batch_size",
		"POSTGRES_DSN":                 "postgres.dsn",
		"RABBIT_HOST":                  "rabbitmq.host",
		"RABBIT_PORT":                  "rabbitmq.port",
		"RABBIT_USER":                  "rabbitmq.username",
		"RABBIT_USERNAME":              "rabbitmq.username",
		"RABBIT_PASSWORD":              "rabbitmq.password",
		"RABBIT_EXCHANGE":              "rabbitmq.exchange",
		"REDIS_HOST":                   "redis.host",
		"REDIS_PORT":                   "redis.port",
		"REDIS_PASSWORD":               "redis.password",
		"REDIS_DB":                     "redis.db",
		"OTEL_COLLECTOR_GRPC_ENDPOINT": "otel.endpoint",
		"OTEL_ENABLED":                 "otel.enabled",
		"LOG_LEVEL":                    "log.level",
		"LOG_DEVELOPMENT":              "log.development",
	}
	if key, ok := mapping[name]; ok {
		return key
	}
	return strings.ToLower(strings.ReplaceAll(name, "__", "."))
}
