package rcpostgres

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/AlexBond702/order-service/internal/app/config/section"
)

type (
	Client struct {
		db  *gorm.DB
		cfg section.RepositoryPostgres
	}
)

func (c *Client) DB() *gorm.DB {
	return c.db
}

func NewConn(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	log.Printf(
		"HOST=%s PORT=%s USER=%s DB=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Name,
	)
	dsn := cfg.DSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	if err := db.Use(otelgorm.NewPlugin(otelgorm.WithDBName(cfg.Name))); err != nil {
		return nil, fmt.Errorf("otelgorm plugin: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetConnMaxLifetime(cfg.ReadTimeout)
	sqlDB.SetConnMaxIdleTime(cfg.WriteTimeout)
	sqlDB.SetMaxOpenConns(10)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	return &Client{
		db:  db,
		cfg: cfg,
	}, nil
}
