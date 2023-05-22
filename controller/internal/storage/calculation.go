package storage

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/kostyay/otel-demo/common/log"
	"gorm.io/gorm/logger"

	"github.com/kostyay/otel-demo/common/version"

	sqlcommentercore "github.com/google/sqlcommenter/go/core"
	gosql "github.com/google/sqlcommenter/go/database/sql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	otelgorm "github.com/kostyay/gorm-opentelemetry"
	"github.com/kostyay/otel-demo/controller/internal/config"
	"github.com/kostyay/otel-demo/controller/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type storage struct {
	db *gorm.DB
}

var timeZoneMatcher = regexp.MustCompile("(time_zone|TimeZone)=(.*?)($|&| )")

const driverName = "pgx"

func dbConnection(cfg *config.Options) (*sql.DB, error) {
	var config *pgx.ConnConfig
	var err error

	config, err = pgx.ParseConfig(cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	config.PreferSimpleProtocol = true

	result := timeZoneMatcher.FindStringSubmatch(cfg.DB.DSN())
	if len(result) > 2 {
		config.RuntimeParams["timezone"] = result[2]
	}

	connStr := stdlib.RegisterConnConfig(config)

	commenterOptions := sqlcommentercore.CommenterOptions{
		Tags: sqlcommentercore.StaticTags{
			Application: version.ServiceName,
			DriverName:  driverName,
		},
		Config: sqlcommentercore.CommenterConfig{
			EnableTraceparent: true,
		},
	}

	return gosql.Open(driverName, connStr, commenterOptions)
}

func New(cfg *config.Options) (*storage, error) {
	dbConn, err := dbConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	log.Info("Connected to database, dsn: " + cfg.DB.DSN())

	pgc := postgres.Config{
		DriverName:           driverName,
		PreferSimpleProtocol: true,
		Conn:                 dbConn,
	}

	log.Info("cfg", pgc)

	db, err := gorm.Open(postgres.New(pgc), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	// Initialize otel plugin with options
	plugin := otelgorm.NewPlugin()
	err = db.Use(plugin)
	if err != nil {
		return nil, fmt.Errorf("unable to use otelgorm plugin: %w", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&domain.Calculation{})
	if err != nil {
		log.WithError(err).Error("unable to migrate database")
	}

	return &storage{db: db}, nil
}

func (s *storage) CreateCalculation(ctx context.Context, owner, expression string) (*domain.Calculation, error) {
	calculation := &domain.Calculation{
		Owner:      owner,
		Expression: expression,
	}
	err := s.db.WithContext(ctx).Create(calculation).Error
	if err != nil {
		return nil, fmt.Errorf("unable to create calculation: %w", err)
	}
	return calculation, nil
}

func (s *storage) GetCalculation(ctx context.Context, id uint) (*domain.Calculation, error) {
	var calculation domain.Calculation
	err := s.db.WithContext(ctx).First(&calculation, id).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find calculation: %w", err)
	}
	return &calculation, nil
}

func (s *storage) GetCalculations(ctx context.Context) ([]*domain.Calculation, error) {
	var calculations []*domain.Calculation
	err := s.db.WithContext(ctx).Find(&calculations).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find calculations: %w", err)
	}
	return calculations, nil
}

func (s *storage) UpdateResult(ctx context.Context, id uint, result float64) error {
	err := s.db.WithContext(ctx).Model(&domain.Calculation{}).Where("id = ?", id).Update("result", result).Update("completed_at", "NOW()").Error
	if err != nil {
		return fmt.Errorf("unable to update calculation: %w", err)
	}
	return nil
}
