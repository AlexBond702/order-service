package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/AlexBond702/order-service/internal/app/client"
	"github.com/AlexBond702/order-service/internal/app/config"
	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
	hhealth "github.com/AlexBond702/order-service/internal/app/handler/health"
	horder "github.com/AlexBond702/order-service/internal/app/handler/order"
	"github.com/AlexBond702/order-service/internal/app/module"
	morder "github.com/AlexBond702/order-service/internal/app/module/order"
	"github.com/AlexBond702/order-service/internal/app/processor"
	rprocessor "github.com/AlexBond702/order-service/internal/app/processor/http"
	"github.com/AlexBond702/order-service/internal/app/processor/monitor"
	"github.com/AlexBond702/order-service/internal/app/repository"
	rcpostgres "github.com/AlexBond702/order-service/internal/app/repository/conn/postgres"
	porder "github.com/AlexBond702/order-service/internal/app/repository/order"
)

type Builder struct {
	cCtx *cli.Context
	ctx  context.Context
	wg   sync.WaitGroup
	err  error
	cfg  config.Config

	connPostgres  *rcpostgres.Client
	orderRepo     repository.Order
	orderModule   module.Order
	orderHandler  rhandler.Order
	healthHandler rhandler.Health

	clientCatalog *client.CatalogClient

	processors []processor.Processor

	chError chan error
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
		repoModule := morder.NewModule(b.orderRepo, b.clientCatalog)
		b.orderModule = repoModule
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
		procHttp := rprocessor.NewHttp(b.healthHandler, b.orderHandler, b.cfg.Processor.WebServer)
		b.processors = append(b.processors, procHttp)
	}, b.healthHandler, b.orderHandler)
}

func (b *Builder) BuildMonitorPrometheus() {
	b.exec(func(b *Builder) {
		if !b.cfg.Monitor.Prometheus.Enabled {
			log.Warn().Msg("Prometheus metrics disabled")
			return
		}
		prometheus := monitor.NewPrometheusObserver()
		b.processors = append(b.processors, prometheus)
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
