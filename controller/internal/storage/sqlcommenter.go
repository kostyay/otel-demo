package storage

import (
	"context"
	"fmt"
	"github.com/kostyay/otel-demo/common/log"
	"go.opentelemetry.io/otel/propagation"
	"gorm.io/gorm"
	"net/url"
)

const (
	callBackBeforeName = "sqlcommenter:before"
	callBackAfterName  = "sqlcommenter:after"
)

type sqlCommenterPlugin struct {
}

func (p *sqlCommenterPlugin) Name() string {
	return "sqlcommenter"
}

func beforeName(name string) string {
	return callBackBeforeName + "_" + name
}

func afterName(name string) string {
	return callBackAfterName + "_" + name
}

func encodeURL(k string) string {
	return url.QueryEscape(k)
}

// ExtractTraceparent extracts the traceparent field using OpenTelemetry library.
func ExtractTraceparent(ctx context.Context) propagation.MapCarrier {
	// Serialize the context into carrier
	textMapPropogator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	carrier := propagation.MapCarrier{}
	textMapPropogator.Inject(ctx, carrier)
	return carrier
}

func (p *sqlCommenterPlugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Create().After("gorm:create").Register(afterName("create"), p.after()); err != nil {
		return err
	}
	if err := db.Callback().Create().Before("gorm:create").Register(beforeName("create"), p.after()); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register(afterName("update"), p.after()); err != nil {
		return err
	}

	if err := db.Callback().Delete().After("gorm:delete").Register(afterName("delete"), p.after()); err != nil {
		return err
	}

	return nil
}

func (p *sqlCommenterPlugin) after() func(tx *gorm.DB) {
	return func(tx *gorm.DB) {
		carrier := ExtractTraceparent(tx.Statement.Context)
		traceID, ok := carrier["traceparent"]
		if !ok {
			log.Info("No traceparent found")
			return
		}

		tx.Statement.SQL.WriteString(fmt.Sprintf(" /*traceparent='%s',framework='gorm'*/", encodeURL(traceID)))
	}
}

func sqlComments(ctx context.Context) string {
	carrier := ExtractTraceparent(ctx)
	traceID, ok := carrier["traceparent"]
	if !ok {
		return ""
	}

	return fmt.Sprintf(" /*traceparent='%s',framework='gorm',route='list',action='list'*/", encodeURL(traceID))
}

func NewSQLCommenterPlugin() *sqlCommenterPlugin {
	return &sqlCommenterPlugin{}
}
