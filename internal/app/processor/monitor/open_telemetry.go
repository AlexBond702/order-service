package mmonitor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/AlexBond702/order-service/internal/app/config/section"
	"github.com/AlexBond702/order-service/internal/app/constant"
	"github.com/AlexBond702/order-service/internal/app/processor"
	"github.com/AlexBond702/order-service/internal/app/util"
)

const (
	initTimeout     = 30 * time.Second
	shutdownTimeout = 5 * time.Second
)

type (
	openTelemetryProc struct {
		traceProvider *sdktrace.TracerProvider
		conn          *grpc.ClientConn
	}

	openTelemetryErrorHandler struct{}
)

func NewOpenTelemetryController(
	ctx context.Context, env string,
	cfg section.MonitorOpenTelemetry,
) (processor.Processor, error) {
	ctx, cancel := context.WithTimeout(ctx, initTimeout)
	defer cancel()

	var p openTelemetryProc

	attributes := []attribute.KeyValue{
		semconv.ServiceName(constant.AppName),
	}
	if env != "" {
		attributes = append(attributes,
			semconv.DeploymentEnvironment(strings.ToLower(env)))
	}
	res, err := resource.New(ctx, resource.WithAttributes(attributes...))
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}
	client, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("create gRPC client for Jaeger: %w", err)
	}
	if err := waitForReady(ctx, client); err != nil {
		return nil, fmt.Errorf("connect to Jaeger at %s: %w", cfg.Address, err)
	}
	p.conn = client
	newconn, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(client))
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}
	cfg.SampleRatio = min(1, max(0, cfg.SampleRatio))
	batchTimeout := sdktrace.WithBatchTimeout(cfg.SendBatchTimeout)
	//nolint:gosec // G115
	maxExport := sdktrace.WithMaxExportBatchSize(int(cfg.MaxBatchSize))
	//nolint:gosec // G115
	maxQueue := sdktrace.WithMaxQueueSize(int(cfg.MaxQueueSize))
	exportTimeout := sdktrace.WithExportTimeout(cfg.ExportTimeout)

	batchProc := sdktrace.NewBatchSpanProcessor(newconn,
		batchTimeout,
		maxExport,
		maxQueue,
		exportTimeout)
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRatio)),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(batchProc))
	p.traceProvider = traceProvider
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetErrorHandler(openTelemetryErrorHandler{})
	log.Info().Str("service", constant.AppName).
		Str("environment", cfg.Address).
		Msg("OpenTelemetry has been initialized")

	return &p, nil
}

func waitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	conn.Connect()
	for {
		if conn.GetState() == connectivity.Ready {
			return nil
		}
		if !conn.WaitForStateChange(ctx, conn.GetState()) {
			return ctx.Err()
		}
	}
}

func (p *openTelemetryProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	processor.WatchForShutdown(ctx, wg, util.CloserFunc(p.shutdown))
}

func (p *openTelemetryProc) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := p.traceProvider.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown trace provider")
	}
	if err := p.conn.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close Jaeger gRPC connection")
	}
	return nil
}

func (openTelemetryErrorHandler) Handle(err error) {
	log.Error().Err(err).Msg("OpenTelemetry error")
}
