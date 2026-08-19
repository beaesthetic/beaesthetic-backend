package messaging

import (
	"context"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/petretiandrea/beaesthetic-backend/rabbitmq"

// StartConsumer extracts the propagated context from an AMQP delivery and starts
// a consumer span as its child.
func StartConsumer(ctx context.Context, queue string, delivery amqp.Delivery) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.operation.name", "process"),
		attribute.String("messaging.destination.name", queue),
	}
	if delivery.Exchange != "" {
		attrs = append(attrs, attribute.String("messaging.rabbitmq.destination.exchange", delivery.Exchange))
	}
	if delivery.RoutingKey != "" {
		attrs = append(attrs, attribute.String("messaging.rabbitmq.destination.routing_key", delivery.RoutingKey))
	}

	return otel.Tracer(instrumentationName).Start(
		otel.GetTextMapPropagator().Extract(ctx, amqpHeaderCarrier(delivery.Headers)),
		"process "+queue,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attrs...),
	)
}

type amqpHeaderCarrier amqp.Table

func (carrier amqpHeaderCarrier) Get(key string) string {
	for header, value := range carrier {
		if !strings.EqualFold(header, key) {
			continue
		}
		switch typed := value.(type) {
		case string:
			return typed
		case []byte:
			return string(typed)
		}
	}
	return ""
}

func (carrier amqpHeaderCarrier) Set(key string, value string) { carrier[key] = value }

func (carrier amqpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(carrier))
	for key := range carrier {
		keys = append(keys, key)
	}
	return keys
}
