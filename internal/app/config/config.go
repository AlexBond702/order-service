package config

import (
	"io"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/AlexBond702/order-service/internal/app/config/section"
	"github.com/AlexBond702/order-service/internal/app/constant"
	msentry "github.com/AlexBond702/order-service/internal/app/processor/monitor/sentry"
)

type (
	Config struct {
		// App        section.App
		Repository section.Repository
		// Broker     section.Broker
		Processor section.Processor
		Monitor   section.Monitor
		Client    section.Client
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
	level, err := zerolog.ParseLevel(Root.Monitor.LogLevel)
	if err != nil {
		log.Warn().Str("log_level", Root.Monitor.LogLevel).Msg("Unknown log level, using debug")
		level = zerolog.DebugLevel
	}
	output := args.Output
	w, ok := msentry.Init(Root.Monitor.Sentry, msentry.Options{
		Release:     constant.Version,
		Environment: Root.Monitor.Environment,
		ServiceName: constant.AppName,
	})
	if ok {
		output = zerolog.MultiLevelWriter(args.Output, w)
	}

	log.Logger = createLogger(level, output)
	log.Info().Msgf("Logger initialized with %s level", level)
	if Root.Monitor.Prometheus.Enabled {
		log.Info().Msg("Prometheus metrics enabled")
	}
}

func createLogger(level zerolog.Level, output io.Writer) zerolog.Logger {
	return zerolog.New(output).Level(level).With().Timestamp().Logger()
}
