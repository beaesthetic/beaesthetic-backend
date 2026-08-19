package messaging

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestInjectAndExtractAMQPContext(t *testing.T) {
	previous := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(previous) })

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{2},
		TraceFlags: trace.FlagsSampled,
	})
	metadata := map[string]string{}
	Inject(trace.ContextWithSpanContext(context.Background(), spanContext), metadata)

	if metadata["traceparent"] == "" {
		t.Fatal("traceparent was not injected")
	}

	headers := amqp.Table{}
	for key, value := range metadata {
		headers[key] = value
	}
	extracted := otel.GetTextMapPropagator().Extract(context.Background(), amqpHeaderCarrier(headers))
	got := trace.SpanContextFromContext(extracted)
	if got.TraceID() != spanContext.TraceID() || got.SpanID() != spanContext.SpanID() || !got.IsRemote() {
		t.Fatalf("unexpected extracted span context: %+v", got)
	}
}
