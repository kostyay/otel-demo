package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/kostyay/otel-demo/common/log"
	"github.com/kostyay/otel-demo/common/otel"
	"github.com/kostyay/otel-demo/common/version"
	"github.com/kostyay/otel-demo/controller/internal/config"
	"github.com/kostyay/otel-demo/controller/internal/handler"
	"github.com/kostyay/otel-demo/controller/internal/math"
	"github.com/kostyay/otel-demo/controller/internal/storage"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func run(ctx context.Context, cfg *config.Options) error {
	db, err := storage.New(cfg)
	if err != nil {
		return fmt.Errorf("unable to initialize storage: %w", err)
	}
	log.Info("storage initialized")

	m, err := math.New(ctx, cfg, db)
	if err != nil {
		return fmt.Errorf("unable to initialize math agent: %w", err)
	}

	log.Info("math agent initialized")

	controller := handler.New(db, m)
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

func main() {
	ctx := context.Background()

	cfg, err := config.Parse()
	if err != nil {
		log.WithError(err).Error("failed to parse config")
		os.Exit(1)
	}

	log.Info("Starting server...")
	tp, err := otel.InitTracing(ctx, otel.Config{
		ProjectID:      cfg.GoogleCloudProject,
		ServiceName:    version.ServiceName,
		ServiceVersion: version.Version,
	})
	if err != nil {
		log.WithError(err).Error("failed to init trace")
		os.Exit(1)
	}

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.WithError(err).Error("failed to shutdown trace")
		}
	}()

	log.Info("Trace initialized")

	if err := run(ctx, cfg); err != nil {
		log.WithError(err).Error("failed to run server")
		os.Exit(1)
	}
}
