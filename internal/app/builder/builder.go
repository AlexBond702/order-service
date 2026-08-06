package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/AlexBond702/order-service/internal/app/client"
	"github.com/AlexBond702/order-service/internal/app/config"
	"github.com/AlexBond702/order-service/internal/app/constant"
	"github.com/AlexBond702/order-service/internal/app/entity"
	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
	ehandler "github.com/AlexBond702/order-service/internal/app/handler/event/order"
	hhealth "github.com/AlexBond702/order-service/internal/app/handler/health"
	horder "github.com/AlexBond702/order-service/internal/app/handler/order"
	"github.com/AlexBond702/order-service/internal/app/module"
	eorder "github.com/AlexBond702/order-service/internal/app/module/event/order"
	morder "github.com/AlexBond702/order-service/internal/app/module/order"
	mmetric "github.com/AlexBond702/order-service/internal/app/monitor/metric"
	"github.com/AlexBond702/order-service/internal/app/processor"
	emonitor "github.com/AlexBond702/order-service/internal/app/processor/event"
	rprocessor "github.com/AlexBond702/order-service/internal/app/processor/http"
	mmonitor "github.com/AlexBond702/order-service/internal/app/processor/monitor"
	"github.com/AlexBond702/order-service/internal/app/repository"
	rcpostgres "github.com/AlexBond702/order-service/internal/app/repository/conn/postgres"
	porder "github.com/AlexBond702/order-service/internal/app/repository/order"
	"github.com/AlexBond702/order-service/internal/app/util"
	"github.com/AlexBond702/order-service/internal/pkg/broker"
	"github.com/AlexBond702/order-service/internal/pkg/broker/codec"
)

type Builder struct {
	cCtx            *cli.Context
	ctx             context.Context
	wg              sync.WaitGroup
	err             error
	cfg             config.Config
	otelServiceName string

	connPostgres  *rcpostgres.Client
	orderRepo     repository.Order
	orderModule   module.Order
	orderHandler  rhandler.Order
	healthHandler rhandler.Health

	clientCatalog *client.CatalogClient

	processors []processor.Processor

	chError chan error

	kafkaClient                *broker.KafkaClient
	busOrderCreated            broker.Bus[entity.EventOrderCreated]
	busOrderDeliveryCalculated broker.Bus[entity.EventOrderDeliveryCalculated]
	updateOrderDeliveryService module.UpdateDelivery
}

func NewBuilder(cCtx *cli.Context) *Builder {
	b := Builder{
		cCtx: cCtx,
		ctx:  context.Background(),
	}
	ctxCancel, cancelFunc := context.WithCancel(context.Background())
	b.ctx = ctxCancel

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go b.waitForSignal(chanSignal, cancelFunc)
	go b.printErrors()

	b.healthHandler = hhealth.NewHealthHandler()
	return &b
}

func (b *Builder) BuildRepoConnPostgres() {
	b.exec(func(b *Builder) {
		configDB := b.cfg.Repository.Postgres
		clientRepo, err := rcpostgres.NewConn(b.ctx, configDB)
		if err != nil {
			b.err = err
			return
		}
		b.connPostgres = clientRepo
	})
}

func (b *Builder) BuildConfig(injectors ...func(c *config.Config)) {
	b.exec(func(b *Builder) {
		b.buildConfig(config.LoadArgs{}, injectors)
	})
}

func (b *Builder) BuildConfigSimple(injectors ...func(c *config.Config)) {
	b.exec(func(b *Builder) {
		b.buildConfig(config.LoadArgs{SkipConfig: true}, injectors)
	})
}

func (b *Builder) BuildRepoOrder(injectors ...func(c *config.Config)) {
	b.exec(func(b *Builder) {
		repoOrder, err := porder.NewOrderRepo(b.ctx, b.connPostgres)
		if err != nil {
			b.err = err
		}
		b.orderRepo = repoOrder
	}, b.connPostgres)
}

func (b *Builder) BuildModuleOrder(injectors ...func(c *config.Config)) {
	b.exec(func(b *Builder) {
		orderMetrics, err := mmetric.NewPrometheusOrder(prometheus.DefaultRegisterer)
		if err != nil {
			b.err = fmt.Errorf("init product metrics: %w", err)
			return
		}
		b.orderModule = morder.NewModule(
			b.orderRepo,
			b.clientCatalog,
			orderMetrics,
			b.busOrderCreated,
		)
	}, b.orderRepo, b.clientCatalog, b.busOrderCreated)
}

func (b *Builder) BuildOrderCalculatedService() {
	b.exec(func(b *Builder) {
		srv := eorder.NewServiceUpdateDelivery(b.orderRepo)
		b.updateOrderDeliveryService = srv
	}, b.orderRepo)
}

func (b *Builder) BuildHandlerHttpOrder() {
	b.exec(func(b *Builder) {
		b.orderHandler = horder.NewHandler(b.orderModule)
	}, b.orderModule)
}

func (b *Builder) BuildCatalogClient() {
	b.exec(func(b *Builder) {
		CatalogAddr := b.cfg.Client.Catalog.GrpcAddr
		clientCatalog, err := client.NewCatalogClient(CatalogAddr)
		if err != nil {
			b.err = fmt.Errorf("failed to create catalog client: %w", err)
			return
		}
		b.clientCatalog = clientCatalog
	}, b.cfg)
}

func (b *Builder) Run() {
	defer func() {
		if b.clientCatalog != nil {
			if err := b.clientCatalog.Close(); err != nil {
				log.Printf("Error closing catalog client: %v", err)
			}
		}
	}()
	if b.err != nil {
		log.Error().Err(b.err).Msg("Failed to initialize application")
	} else {
		log.Info().Msg("Application is initialized")
	}
	defer log.Info().Msg("Application is completed, GoodBye!")

	for _, proc := range b.processors {
		proc.StartAsync(b.ctx, &b.wg)
	}
	b.wg.Wait()
}

func (b *Builder) BuildProcHttp() {
	b.exec(func(b *Builder) {
		procHttp := rprocessor.NewHttp(b.otelServiceName, b.healthHandler, b.orderHandler, b.cfg.Processor.WebServer)
		b.processors = append(b.processors, procHttp)
	}, b.healthHandler, b.orderHandler)
}

func (b *Builder) BuildProcHttpAdmin() {
	b.exec(func(b *Builder) {
		proc := rprocessor.NewHttpAdmin(b.otelServiceName, b.healthHandler, b.cfg.Processor.WebServer)
		b.processors = append(b.processors, proc)
	}, b.healthHandler)
}

func (b *Builder) BuildMonitorPrometheus() {
	b.exec(func(b *Builder) {
		if !b.cfg.Monitor.Prometheus.Enabled {
			log.Warn().Msg("Prometheus metrics disabled")
			return
		}
		prometheus := mmetric.NewPrometheusObserver()
		b.processors = append(b.processors, prometheus)
	})
}

func (b *Builder) BuildMonitorOpenTelemetry() {
	cfg := b.cfg.Monitor.OpenTelemetry
	if !cfg.Enabled {
		log.Warn().Msg("OpenTelemetry is disabled by config")
		return
	}
	b.exec(func(b *Builder) {
		proc, err := mmonitor.NewOpenTelemetryController(b.ctx, b.cfg.Monitor.Environment, cfg)
		if err != nil {
			b.err = fmt.Errorf("init OpenTelemetry: %w", err)
			return
		}
		b.processors = append(b.processors, proc)
		b.otelServiceName = constant.AppName
	})
}

func (b *Builder) buildConfig(args config.LoadArgs, injectors []func(*config.Config)) {
	args.Output = b.cCtx.App.Writer
	args.EnableSimpleLog = b.cCtx.Bool("no json")

	config.Load(args)

	for i, injector := range injectors {
		if injectors[i] != nil {
			injector(&config.Root)
		}
	}
	b.cfg = config.Root
}

func (b *Builder) BuildBrokerKafka() {
	b.exec(func(b *Builder) {
		cfg := b.cfg.Broker.Kafka
		kafkaClient, err := broker.NewKafkaClient(broker.KafkaConfig{
			Addresses:     cfg.Addresses,
			ConsumerGroup: cfg.ConsumerGroup,
			ClientID:      cfg.ClientID,
		})
		if err != nil {
			b.err = fmt.Errorf("failed to create kafka client: %w", err)
			return
		}
		b.kafkaClient = kafkaClient
		b.processors = append(b.processors, processor.ProcessorFunc(func(ctx context.Context, wg *sync.WaitGroup) {
			processor.WatchForShutdown(ctx, wg, util.CloserFunc(b.kafkaClient.Close))
		}))
	})
}

func (b *Builder) BuildBusOrderCreated() {
	b.exec(func(b *Builder) {
		cfg := b.cfg.Broker.Kafka
		topic := cfg.ModelOrder.Created.Topic
		group := broker.Coalesce(cfg.ModelOrder.Created.ConsumerGroup, cfg.ConsumerGroup)

		bus, err := broker.NewBus[entity.EventOrderCreated](
			b.kafkaClient,
			codec.NewCodecJson[entity.EventOrderCreated](),
			topic,
			group)
		if err != nil {
			b.err = fmt.Errorf("failed to create bus order.created: %w", err)
			return
		}
		b.busOrderCreated = bus
	}, b.kafkaClient)
}

func (b *Builder) BuildConsumerOrderDeliveryCalculated() {
	b.exec(func(b *Builder) {
		cfg := b.cfg.Broker.Kafka
		topic := cfg.ModelOrder.Delivery.Topic
		group := broker.Coalesce(cfg.ModelOrder.Delivery.ConsumerGroup, cfg.ConsumerGroup)

		bus, err := broker.NewBus[entity.EventOrderDeliveryCalculated](
			b.kafkaClient,
			codec.NewCodecJson[entity.EventOrderDeliveryCalculated](),
			topic,
			group,
		)
		if err != nil {
			b.err = err
			return
		}
		b.busOrderDeliveryCalculated = bus

		handler := ehandler.NewHandlerOrderDelivery(b.updateOrderDeliveryService)
		b.processors = append(b.processors, emonitor.NewProc(handler, b.busOrderDeliveryCalculated))
	}, b.kafkaClient, b.updateOrderDeliveryService)
}

func (b *Builder) exec(cb func(b *Builder), requiredArgs ...any) {
	if b.err != nil {
		return
	}

	for i, requiredArg := range requiredArgs {
		rv := reflect.ValueOf(requiredArg)
		if !rv.IsValid() {
			b.err = fmt.Errorf("BUG: required argument #%d is nil (check dependencies)", i)
			return
		}
		if rv.Type().Kind() == reflect.Struct || !rv.IsZero() {
			continue
		}
		b.err = fmt.Errorf("BUG: required %s, but empty", rv.Type().String())
		return
	}
	cb(b)
}

func (b *Builder) waitForSignal(sig chan os.Signal, cancelFunc func()) {
	sigValue := <-sig
	log.Printf("Signal arrive: %T", sigValue)
	cancelFunc()
}

func (b *Builder) printErrors() {
	for chErr := range b.chError {
		log.Error().Err(chErr)
	}
}
