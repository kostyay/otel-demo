package main

import (
	"context"
	"fmt"
	"github.com/kostyay/otel-demo/common/log"
	"github.com/kostyay/otel-demo/controller/internal/config"
	"github.com/kostyay/otel-demo/controller/internal/handler"
	"github.com/kostyay/otel-demo/controller/internal/math"
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
	log.Info("storage initialized")

	m, err := math.New(context.Background(), cfg, db)

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
