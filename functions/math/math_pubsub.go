// Package helloworld provides a set of Cloud Functions samples.
package math

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kostyay/otel-demo/common/log"
	common_otel "github.com/kostyay/otel-demo/common/otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/maja42/goval"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.15.0"
	"go.opentelemetry.io/otel/trace"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"

	pb "github.com/kostyay/otel-demo/controller/api/calculator/v1"
)

func init() {
	_, err := common_otel.InitTracing(context.Background(), common_otel.Config{
		ProjectID:      os.Getenv("GOOGLE_CLOUD_PROJECT"),
		ServiceName:    "math",
		ServiceVersion: "0.0.1",
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize tracing")
		os.Exit(1)
	}
	functions.CloudEvent("calculateExpression", calculateExpression)
}

// MessagePublishedData contains the full Pub/Sub message
// See the documentation for more details:
// https://cloud.google.com/eventarc/docs/cloudevents#pubsub
type MessagePublishedData struct {
	Message PubSubMessage
}

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data       []byte            `json:"data"`
	Attributes map[string]string `json:"attributes"`
	ID         string            `json:"messageId"`
}

func spanFromPubsubMessage(ctx context.Context, tracer trace.Tracer, topicID string, msg PubSubMessage) (context.Context, trace.Span) {
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

// helloPubSub consumes a CloudEvent message and extracts the Pub/Sub message.
func calculateExpression(ctx context.Context, e event.Event) error {
	var err error
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %v", err)
	}

	ctx, span := spanFromPubsubMessage(ctx, common_otel.Tracer(), "math-topic", msg.Message)
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()

	logger := log.WithContext(ctx)

	var calculation pb.Calculation

	err = json.Unmarshal(msg.Message.Data, &calculation)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}

	span.SetAttributes(attribute.String("owner", calculation.GetOwner()), attribute.String("expression", calculation.GetExpression()))

	logger.Infof("Calculation: Owner: %s; Expression: %s; Attributes: %v", calculation.GetOwner(), calculation.GetExpression(), msg.Message.Attributes)

	span.AddEvent("evaluating expression")
	eval := goval.NewEvaluator()
	result, err := eval.Evaluate(calculation.GetExpression(), nil, nil)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to evaluate expression")
		return err
	}

	logger.Infof("Result: %f", result)

	return nil
}
