package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.15.0"
	"go.opentelemetry.io/otel/trace"
)

func BeforePublishMessage(ctx context.Context, tracer trace.Tracer, topicID string, msg *pubsub.Message) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			// customizable span attributes
			semconv.MessagingSystemKey.String("pubsub"),
			semconv.MessagingDestinationKey.String(topicID),
			semconv.MessagingDestinationKindTopic,
		),
	}
	ctx, span := tracer.Start(ctx, fmt.Sprintf("%s send", topicID), opts...)
	if msg.Attributes == nil {
		msg.Attributes = make(map[string]string)
	}
	// propagate Span across process boundaries
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(msg.Attributes))
	return ctx, span
}
func AfterPublishMessage(span trace.Span, messageID string, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetAttributes(semconv.MessagingMessageIDKey.String(messageID))
	}
}
