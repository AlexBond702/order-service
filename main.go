package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/AlexBond702/order-service/internal/app/config"
	rhealth "github.com/AlexBond702/order-service/internal/app/handler/health"
	"github.com/AlexBond702/order-service/internal/app/processor/http"
	"github.com/AlexBond702/order-service/internal/app/repository/postgres"
)

func main() {
	config.Load(config.LoadArgs{
		Output:          os.Stdout,
		EnableSimpleLog: false,
		SkipConfig:      false,
	})
	cfg := config.Root
	log.Print("Configuration loaded successfully")

	pgClient, err := postgres.NewConn(context.Background(), cfg.Repository.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed connect to PostgresSQL")
	}
	log.Printf("Connected to database: %s", cfg.Repository.Postgres.Name)
	_ = pgClient

	healthHandler := rhealth.NewHealthHandler()

	server := http.NewHttp(healthHandler, cfg.Processor.WebServer)

	if err := server.Run(); err != nil {
		log.Fatal().Err(err).Msg("HTTP server failed")
	}
}
