package http

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AlexBond702/order-service/internal/app/config/section"
	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
)

type Processor struct {
	server *http.Server
}

func NewHttp(hHealth rhandler.Health,
	cfg section.ProcessorWebServer,
) *Processor {
	router := gin.Default()

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

func (p *Processor) Run() error {
	log.Println("Starting HTTP Server...")
	return p.server.ListenAndServe()
}

func (p *Processor) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return p.server.Shutdown(ctx)
}
