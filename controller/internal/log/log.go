package log

import (
	"context"
	"github.com/kostyay/otel-demo/controller/internal/config"
	"github.com/kostyay/zapdriver"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"log"
)

const (
	traceKey        = "logging.googleapis.com/trace"
	spanKey         = "logging.googleapis.com/spanId"
	traceSampledKey = "logging.googleapis.com/trace_sampled"
	errorKey        = "err"
)

type Logger struct {
	logger *zap.SugaredLogger
}

var globalLog *Logger

func init() {
	zapconf := zapdriver.NewProductionConfig()

	l, err := zapconf.Build(
		zapdriver.WrapCore(
			zapdriver.ReportAllErrors(true),
			zapdriver.ServiceName(config.ServiceName),
			zapdriver.ServiceVersion(config.ServiceVersion),
		),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	globalLog = &Logger{logger: l.Sugar()}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	span := trace.SpanFromContext(ctx).SpanContext()

	return &Logger{
		logger: l.logger.With(
			traceKey, span.TraceID().String(),
			spanKey, span.SpanID().String(),
			traceSampledKey, span.IsSampled())}
}

func (l *Logger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func WithContext(ctx context.Context) *Logger {
	return globalLog.WithContext(ctx)
}

func Debug(args ...interface{}) {
	globalLog.Debug(args...)
}

func Info(args ...interface{}) {
	globalLog.Info(args...)
}

func Fatalf(format string, args ...interface{}) {
	globalLog.logger.Fatalf(format, args...)
}

func WithError(err error) *Logger {
	return &Logger{logger: globalLog.logger.With(errorKey, err.Error())}
}
