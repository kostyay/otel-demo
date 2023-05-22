package math

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/kostyay/otel-demo/common/log"
	otelcommon "github.com/kostyay/otel-demo/common/otel"
	otelpubsub "github.com/kostyay/otel-demo/common/otel/pubsub"
	pb "github.com/kostyay/otel-demo/controller/api/calculator/v1"
	"github.com/kostyay/otel-demo/controller/internal/config"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Storage interface {
	UpdateResult(ctx context.Context, id uint, result float64) error
}

type handler struct {
	requestTopic *pubsub.Topic
	responseSub  *pubsub.Subscription
	client       *pubsub.Client
	storage      Storage
}

func New(ctx context.Context, cfg *config.Options, storage Storage) (*handler, error) {
	requestClient, err := pubsub.NewClient(ctx, cfg.GoogleCloudProject)
	if err != nil {
		return nil, fmt.Errorf("unable to create pubsub client: %w", err)
	}

	result := &handler{
		client:       requestClient,
		requestTopic: requestClient.Topic(cfg.MathRequestTopic),
		responseSub:  requestClient.Subscription(cfg.MathResultSubscription),
		storage:      storage,
	}

	// Create the requestTopic if it doesn't exist.
	exists, err := result.requestTopic.Exists(ctx)
	if !exists || err != nil {
		requestClient.Close()
		return nil, fmt.Errorf("unable to check request topic existence: %w", err)
	}

	exists, err = result.responseSub.Exists(ctx)
	if !exists || err != nil {
		requestClient.Close()
		return nil, fmt.Errorf("unable to check response subscription existence: %w", err)
	}

	go func() {
		err := result.responseSub.Receive(ctx, otelpubsub.WrapPubSubHandlerWithTelemetry(otelcommon.Tracer(), cfg.MathResultSubscription, result.handleMathResult))
		if err != nil {
			log.WithError(err).Error("unable to receive pubsub messages")
		}
	}()

	return result, nil
}

func (h *handler) handleMathResult(ctx context.Context, msg *pubsub.Message) {
	var calculation pb.Calculation

	var err error
	span := trace.SpanFromContext(ctx)

	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	logger := log.WithContext(ctx)

	err = json.Unmarshal(msg.Data, &calculation)
	if err != nil {
		logger.WithError(err).Error("unable to unmarshal calculation")
		return
	}

	err = h.storage.UpdateResult(ctx, uint(calculation.Id), calculation.Result)
	if err != nil {
		logger.WithError(err).Error("unable to update result")
		return
	}

	span.AddEvent("result updated")

	msg.Ack()
}

func (h *handler) Calculate(ctx context.Context, calculation *pb.Calculation) error {
	expression, err := json.Marshal(calculation)
	if err != nil {
		return fmt.Errorf("unable to marshal calculation: %w", err)
	}

	msg := &pubsub.Message{
		Data: expression,
	}
	// Create a new span
	ctx, span := otelpubsub.BeforePublishMessage(ctx, otelcommon.Tracer(), h.requestTopic.String(), msg)
	defer span.End()

	result, err := h.requestTopic.Publish(ctx, msg).Get(ctx)
	otelpubsub.AfterPublishMessage(span, result, err)
	if err != nil {
		return fmt.Errorf("unable to publish message: %w", err)
	}

	return nil
}

func (h *handler) Close() error {
	return h.client.Close()
}
