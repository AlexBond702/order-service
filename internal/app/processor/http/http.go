package processor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/AlexBond702/order-service/internal/app/config/section"
	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
	"github.com/AlexBond702/order-service/internal/app/processor"
	"github.com/AlexBond702/order-service/internal/app/util"
	"github.com/AlexBond702/order-service/internal/pkg/http/httph"
	"github.com/AlexBond702/order-service/internal/pkg/http/mzerolog"
)

type Processor struct {
	server *http.Server
	processor.Processor
}

func NewHttp(hHealth rhandler.Health,
	cfg section.ProcessorWebServer,
) *Processor {
	router := gin.Default()
	router.Use(
		httph.NewErrorMiddleware(),
		mzerolog.NewMiddleware(
			mzerolog.WithSkipper(util.IsFilteredHttpRoute)),
		makeErrorMiddleware(),
	)

	GenericRegHealthCheck(router, hHealth)

	logRoutes(router)

	addr := fmt.Sprintf("%s:%v", cfg.Host, cfg.ListenPort)

	svr := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}
	log.Printf("HTTP server configured on %s", addr)

	return &Processor{
		server: svr,
	}
}

func (p *Processor) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", p.server.Addr)
	if err != nil {
		log.Fatal().Err(err).Str("listen_addr", p.server.Addr).
			Msg("Failed to start listening TCP addr for HTTP server")
		return
	}
	log.Info().Str("listen_addr", p.server.Addr).
		Msg("Listening of TCP addr for HTTP server has been started")
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.serve(l)
	}()
	go processor.WatchForShutdown(ctx, wg, util.CloserFunc(l.Close))
	go processor.WatchForShutdown(ctx, wg, util.NewCloserContextFunc(p.server.Shutdown, context.Background(), time.Second*5))
	wg.Wait()
}

func (p *Processor) serve(l net.Listener) {
	log.Info().Str("addr", p.server.Addr).Msg("Starting HTTP server")
	_ = p.server.Serve(l)
}

func (p *Processor) Run() error {
	log.Print("Starting HTTP Server...")
	return p.server.ListenAndServe()
}

func (p *Processor) Shutdown(ctx context.Context) error {
	log.Print("Shutting down HTTP server...")
	return p.server.Shutdown(ctx)
}
