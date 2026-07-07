package rabbitmq

import (
	"testing"

	"github.com/beaesthetic/scheduler/internal/config"
)

func TestAMQPURLUsesDefaultVHost(t *testing.T) {
	got := amqpURL(config.RabbitMQConfig{
		Host:     "rabbitmq",
		Port:     5672,
		Username: "user",
		Password: "pass",
		VHost:    "/",
	})

	want := "amqp://user:pass@rabbitmq:5672/"
	if got != want {
		t.Fatalf("amqpURL() = %q, want %q", got, want)
	}
}

func TestAMQPURLUsesConfiguredVHost(t *testing.T) {
	got := amqpURL(config.RabbitMQConfig{
		Host:     "rabbitmq",
		Port:     5672,
		Username: "user",
		Password: "pass",
		VHost:    "dev",
	})

	want := "amqp://user:pass@rabbitmq:5672/dev"
	if got != want {
		t.Fatalf("amqpURL() = %q, want %q", got, want)
	}
}
