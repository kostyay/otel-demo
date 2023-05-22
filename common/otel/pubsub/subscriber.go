package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.15.0"
	"go.opentelemetry.io/otel/trace"
)

type PubSubHandler = func(context.Context, *pubsub.Message)

func WrapPubSubHandlerWithTelemetry(tracer trace.Tracer, topicID string, handler PubSubHandler) PubSubHandler {
	return func(ctx context.Context, msg *pubsub.Message) {
		// create span
		ctx, span := beforePubSubHandlerInvoke(ctx, tracer, topicID, msg)
		defer span.End()
		// call actual handler function
		handler(ctx, msg)
	}
}
func beforePubSubHandlerInvoke(ctx context.Context, tracer trace.Tracer, topicID string, msg *pubsub.Message) (context.Context, trace.Span) {
	if msg.Attributes != nil {
		// extract propagated span
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.MapCarrier(msg.Attributes))
	}
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			//customizable attributes
			semconv.FaaSTriggerPubsub,
			semconv.MessagingSystemKey.String("pubsub"),
			semconv.MessagingDestinationKey.String(topicID),
			semconv.MessagingDestinationKindTopic,
			semconv.MessagingOperationProcess,
			semconv.MessagingMessageIDKey.String(msg.ID),
		),
	}
	return tracer.Start(ctx, fmt.Sprintf("%s process", topicID), opts...)
}
