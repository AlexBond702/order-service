package config

import (
	"io"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/AlexBond702/order-service/internal/app/config/section"
)

type (
	Config struct {
		// App        section.App
		Repository section.Repository
		// Broker     section.Broker
		Processor section.Processor
		Monitor   section.Monitor
	}
	LoadArgs struct {
		Output          io.Writer `json:"-"`
		EnableSimpleLog bool
		SkipConfig      bool
	}
)

var Root Config

func Load(args LoadArgs) {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.MessageFieldName = "msg"
	zerolog.TimeFieldFormat = time.RFC3339

	if args.EnableSimpleLog {
		args.Output = zerolog.ConsoleWriter{Out: args.Output}
	}
	log.Logger = createLogger(zerolog.DebugLevel, args.Output)
	log.Debug().Msg("Logger initialized with Debug level")

	if args.SkipConfig {
		log.Debug().Msg("Config loading skipped")
		return
	}

	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("Error loading .env file")
	}
	if err := envconfig.Process("APP", &Root); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse")
	}
	level := zerolog.DebugLevel
	if levelLog, err := zerolog.ParseLevel(Root.Monitor.LogLevel); err == nil {
		level = levelLog
	} else {
		log.Warn().Msgf("failed parsing level: %s", level)
	}

	log.Logger = createLogger(level, args.Output)
	log.Info().Msgf("Logger initialized with %s level", level)
}

func createLogger(level zerolog.Level, output io.Writer) zerolog.Logger {
	return zerolog.New(output).Level(level).With().Timestamp().Logger()
}
