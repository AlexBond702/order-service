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

	"github.com/AlexBond702/order-service/internal/app/config"
	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
	hhealth "github.com/AlexBond702/order-service/internal/app/handler/health"
	"github.com/AlexBond702/order-service/internal/app/processor"
	rprocessor "github.com/AlexBond702/order-service/internal/app/processor/http"
	rcpostgres "github.com/AlexBond702/order-service/internal/app/repository/conn/postgres"
)

type Builder struct {
	cCtx *cli.Context
	ctx  context.Context
	wg   sync.WaitGroup
	err  error
	cfg  config.Config

	connPostgres *rcpostgres.Client
	// orderRepo    repository.Order

	healthHandler rhandler.Health

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
		client, err := rcpostgres.NewConn(b.ctx, configDB)
		if err != nil {
			b.err = err
			return
		}
		b.connPostgres = client
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

func (b *Builder) Run() {
	if b.err != nil {
		log.Fatal().Err(b.err).Msg("Failed to initialize application")
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
		procHttp := rprocessor.NewHttp(b.healthHandler, b.cfg.Processor.WebServer)
		b.processors = append(b.processors, procHttp)
	}, b.healthHandler)
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
