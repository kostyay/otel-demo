package storage

import (
	"context"
	"fmt"
	"github.com/kostyay/otel-demo/common/log"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	otelgorm "github.com/kostyay/gorm-opentelemetry"
	"github.com/kostyay/otel-demo/controller/internal/config"
	"github.com/kostyay/otel-demo/controller/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type storage struct {
	db *gorm.DB
}

func New(cfg *config.Options) (*storage, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", cfg.DB.InstanceConnectionName, cfg.DB.User, cfg.DB.Name, cfg.DB.Password)
	log.Infof("Connecting to database, dsn=%s", dsn)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: "cloudsqlpostgres",
		DSN:        dsn,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	// Initialize otel plugin with options
	plugin := otelgorm.NewPlugin()
	err = db.Use(plugin)
	if err != nil {
		return nil, fmt.Errorf("unable to use otelgorm plugin: %w", err)
	}

	//sqlcPlugin := NewSQLCommenterPlugin()
	//err = db.Use(sqlcPlugin)
	//if err != nil {
	//	return nil, fmt.Errorf("unable to use sqlc plugin: %w", err)
	//}

	// Migrate the schema
	err = db.AutoMigrate(&domain.Calculation{})
	if err != nil {
		panic(err.Error())
	}

	return &storage{db: db}, nil
}

func (s *storage) CreateCalculation(ctx context.Context, owner, expression string) (*domain.Calculation, error) {
	calculation := &domain.Calculation{
		Owner:      owner,
		Expression: expression,
	}
	err := s.db.WithContext(ctx).Debug().Create(calculation).Error
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
	err := s.db.WithContext(ctx).Debug().Raw("SELECT * FROM calculations ORDER BY created_at DESC " + sqlComments(ctx)).Scan(&calculations).Error
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
