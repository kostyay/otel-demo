package math

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	pb "github.com/kostyay/otel-demo/controller/api/calculator/v1"
	"github.com/kostyay/otel-demo/controller/internal/config"
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

	return result, nil
}

func (h *handler) Calculate(ctx context.Context, calculation *pb.Calculation) error {
	expression, err := json.Marshal(calculation)
	if err != nil {
		return fmt.Errorf("unable to marshal calculation: %w", err)
	}

	result := h.requestTopic.Publish(ctx, &pubsub.Message{
		Data: expression,
	})

	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("unable to publish message: %w", err)
	}

	return nil
}

func (h *handler) Close() error {
	return h.client.Close()
}
