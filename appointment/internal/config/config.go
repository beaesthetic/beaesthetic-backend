package config

import (
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const delim = "."

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
	k := koanf.New(delim)
	for key, value := range map[string]interface{}{
		"app.name":                            "beaesthetic-agenda",
		"app.env":                             "dev",
		"http.addr":                           ":8080",
		"mongo.database":                      "appointments-v2",
		"reminder.trigger_before":             "24h",
		"reminder.immediate_send_threshold":   "2m",
		"reminder.no_send_threshold":          "30m",
		"rabbitmq.url":                        "amqp://guest:guest@localhost:5672/",
		"rabbitmq.scheduler_queue":            "reminders",
		"rabbitmq.notification_confirm_queue": "NotificationConfirmQueue",
	} {
		if err := k.Set(key, value); err != nil {
			return Config{}, err
		}
	}
	if strings.TrimSpace(envFile) != "" {
		if err := k.Load(file.Provider(envFile), dotenv.ParserEnv("", delim, normalizeEnvKey)); err != nil {
			return Config{}, err
		}
	}
	if err := k.Load(env.Provider("", delim, normalizeEnvKey), nil); err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func normalizeEnvKey(key string) string { return strings.ToLower(strings.ReplaceAll(key, "__", delim)) }
