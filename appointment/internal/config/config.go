package config

import (
	"strings"
	"time"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

const ENV_PREFIX = "ENV_"

type Config struct {
	App      AppConfig      `koanf:"app"`
	HTTP     HTTPConfig     `koanf:"http"`
	Postgres PostgresConfig `koanf:"postgres"`
	Mongo    MongoConfig    `koanf:"mongo"`
	Remote   RemoteConfig   `koanf:"remote"`
	Reminder ReminderConfig `koanf:"reminder"`
	RabbitMQ RabbitMQConfig `koanf:"rabbitmq"`
}

type AppConfig struct {
	Name string `koanf:"name"`
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
type RemoteConfig struct {
	CustomerURL     string `koanf:"customer_url"`
	SchedulerURL    string `koanf:"scheduler_url"`
	NotificationURL string `koanf:"notification_url"`
}
type RabbitMQConfig struct {
	URL                      string `koanf:"url"`
	SchedulerQueue           string `koanf:"scheduler_queue"`
	NotificationConfirmQueue string `koanf:"notification_confirm_queue"`
}

type ReminderConfig struct {
	TriggerBefore          time.Duration `koanf:"trigger_before"`
	ImmediateSendThreshold time.Duration `koanf:"immediate_send_threshold"`
	NoSendThreshold        time.Duration `koanf:"no_send_threshold"`
}

func Load(envFile string) (Config, error) {
	k := koanf.New(".")

	if err := k.Load(env.Provider(".", env.Opt{
		Prefix:        ENV_PREFIX,
		TransformFunc: trasnformFunction,
	}), nil); err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func StripUnderscore(s, newDelimiter string) string {
	var new = strings.Replace(s, "__", "@", -1)
	new = strings.Replace(new, "_", newDelimiter, -1)
	new = strings.Replace(new, "@", "_", -1)
	return new
}

func trasnformFunction(k, v string) (string, any) {
	// convert to lowercase and replace underscores with dots, except for double underscores which are replaced with a single underscore
	k = StripUnderscore(strings.ToLower(strings.TrimPrefix(k, ENV_PREFIX)), ".")
	// split for array, like "ENV_MY_ARRAY=val1 val2 val3" into []string{"val1", "val2", "val3"}
	if strings.Contains(v, " ") {
		return k, strings.Split(v, " ")
	}
	return k, v
}
