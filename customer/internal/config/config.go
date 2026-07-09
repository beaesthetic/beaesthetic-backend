package config

import (
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	App      AppConfig      `koanf:"app"`
	HTTP     HTTPConfig     `koanf:"http"`
	Postgres PostgresConfig `koanf:"postgres"`
	Mongo    MongoConfig    `koanf:"mongo"`
	Redis    RedisConfig    `koanf:"redis"`
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
	URI      string `koanf:"uri"`
	Database string `koanf:"database"`
}

type RedisConfig struct {
	URI                string        `koanf:"uri"`
	CustomersTTL       time.Duration `koanf:"customers_ttl"`
	CustomersSearchTTL time.Duration `koanf:"customers_search_ttl"`
}

const (
	keyDelimiter = "."
	envPrefix    = "ENV_"
)

func Load(envFile string) (Config, error) {
	k := koanf.New(keyDelimiter)
	_ = k.Set("app.name", "beaesthetic-customers")
	_ = k.Set("app.env", "local")
	_ = k.Set("http.addr", ":8080")
	_ = k.Set("mongo.uri", "mongodb://localhost:27017")
	_ = k.Set("mongo.database", "customer")
	_ = k.Set("redis.uri", "redis://localhost:6379")
	_ = k.Set("redis.customers_ttl", "1h")
	_ = k.Set("redis.customers_search_ttl", "5m")
	if envFile != "" {
		if err := k.Load(file.Provider(envFile), dotenv.ParserEnv(envPrefix, keyDelimiter, normalizeEnvKey)); err != nil {
			return Config{}, err
		}
	}
	if err := k.Load(env.Provider(envPrefix, keyDelimiter, normalizeEnvKey), nil); err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func normalizeEnvKey(key string) string {
	key = strings.TrimPrefix(key, envPrefix)
	return stripUnderscore(strings.ToLower(key), keyDelimiter)
}

func stripUnderscore(value, newDelimiter string) string {
	value = strings.ReplaceAll(value, "__", "@")
	value = strings.ReplaceAll(value, "_", newDelimiter)
	return strings.ReplaceAll(value, "@", "_")
}
