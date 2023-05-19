package main

import (
	"context"
	"fmt"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/kostyay/otel-demo/controller/internal/config"
	"github.com/kostyay/otel-demo/controller/internal/handler"
	"github.com/kostyay/otel-demo/controller/internal/log"
	"github.com/kostyay/otel-demo/controller/internal/storage"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
	"os"
)

func run() error {
	cfg, err := config.Parse()
	if err != nil {
		log.WithError(err).Error("failed to parse config")
		return err
	}

	db, err := storage.New(cfg)
	if err != nil {
		return fmt.Errorf("unable to initialize storage: %w", err)
	}

	controller := handler.New(db)
	mux := http.NewServeMux()
	// The generated constructors return a path and a plain net/http
	// handler.
	controller.Register(mux)
	return http.ListenAndServe(
		cfg.ListenAddr,
		// For gRPC clients, it's convenient to support HTTP/2 without TLS. You can
		// avoid x/net/http2 by using http.ListenAndServeTLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
}

func initTrace() error {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Identify your application using resource detection
	res, err := resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
		),
	)
	if err != nil {
		return err
	}

	// Create trace provider with the exporter.
	//
	// By default it uses AlwaysSample() which samples all traces.
	// In a production environment or high QPS setup please use
	// probabilistic sampling.
	// Example:
	//   tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.0001)), ...)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return nil
}

func main() {
	log.Info("Starting server...")

	if err := initTrace(); err != nil {
		log.WithError(err).Error("failed to init trace")
		os.Exit(1)
	}

	log.Info("Trace initialized")

	if err := run(); err != nil {
		log.WithError(err).Error("failed to run server")
		os.Exit(1)
	}
}
