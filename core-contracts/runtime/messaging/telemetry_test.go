package messaging

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestExtractAMQPContext(t *testing.T) {
	previous := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(previous) })

	headers := amqp.Table{}
	headers["traceparent"] = "00-00000000000000000000000000000001-0000000000000002-01"
	extracted := otel.GetTextMapPropagator().Extract(context.Background(), amqpHeaderCarrier(headers))
	got := trace.SpanContextFromContext(extracted)
	if got.TraceID().String() != "00000000000000000000000000000001" || got.SpanID().String() != "0000000000000002" || !got.IsRemote() {
		t.Fatalf("unexpected extracted span context: %+v", got)
	}
}
