package math

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/kostyay/otel-demo/common/log"
	otelcommon "github.com/kostyay/otel-demo/common/otel"
	otelpubsub "github.com/kostyay/otel-demo/common/otel/pubsub"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"os"

	"github.com/maja42/goval"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.15.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"

	pb "github.com/kostyay/otel-demo/controller/api/calculator/v1"
)

const (
	resultTopic = "math-result-topic"
)

var googleCloudProject = os.Getenv("GOOGLE_CLOUD_PROJECT")

func init() {
	_, err := otelcommon.InitTracing(context.Background(), otelcommon.Config{
		ProjectID:      googleCloudProject,
		ServiceName:    "math-cloud-function",
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

	ctx, span := spanFromPubsubMessage(ctx, otelcommon.Tracer(), "math-topic", msg.Message)
	defer func(err1 *error) {
		if err != nil {
			log.WithError(*err1).Error("Failed to process message")
			span.RecordError(*err1)
			span.SetStatus(codes.Error, (*err1).Error())
		}
		span.End()
	}(&err)

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
	//
	if intResult, ok := result.(int); ok {
		calculation.Result = float64(intResult)
	} else if floatResult, ok := result.(float64); ok {
		calculation.Result = floatResult
	} else {
		calculation.Result = 0
	}

	err = sendResult(ctx, &calculation)
	if err != nil {
		logger.WithError(err).Error("Failed to send result")
		return err
	}

	span.AddEvent("result sent")

	return nil
}

func sendResult(ctx context.Context, calc *pb.Calculation) error {
	client, err := pubsub.NewClient(ctx, googleCloudProject)
	if err != nil {
		return fmt.Errorf("unable to create pubsub client: %w", err)
	}
	defer client.Close()

	topic := client.Topic(resultTopic)
	exists, err := topic.Exists(ctx)
	if err != nil || !exists {
		return fmt.Errorf("unable to get topic: %w", err)
	}

	respJson, err := json.Marshal(calc)
	if err != nil {
		return fmt.Errorf("unable to marshal calculation: %w", err)
	}

	msg := &pubsub.Message{
		Data: respJson,
	}

	// Create a new span
	ctx, span := otelpubsub.BeforePublishMessage(ctx, otelcommon.Tracer(), resultTopic, msg)
	defer span.End()

	result, err := topic.Publish(ctx, msg).Get(ctx)
	otelpubsub.AfterPublishMessage(span, result, err)
	if err != nil {
		return fmt.Errorf("unable to publish message: %w", err)
	}

	return nil

}
