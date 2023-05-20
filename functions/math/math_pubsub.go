// Package helloworld provides a set of Cloud Functions samples.
package math

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kostyay/otel-demo/common/log"
	"github.com/kostyay/otel-demo/common/otel"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"

	pb "github.com/kostyay/otel-demo/controller/api/calculator/v1"
)

func init() {
	_, err := otel.InitTracing(context.Background(), otel.Config{
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
	Data []byte `json:"data"`
}

// helloPubSub consumes a CloudEvent message and extracts the Pub/Sub message.
func calculateExpression(ctx context.Context, e event.Event) error {
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %v", err)
	}

	var calculation pb.Calculation

	err := json.Unmarshal(msg.Message.Data, &calculation)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	log.WithContext(ctx).Infof("Calculation: %-v", calculation)

	return nil
}
