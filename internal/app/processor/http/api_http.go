package processor

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/AlexBond702/order-service/internal/app/config/section"
	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
	"github.com/AlexBond702/order-service/internal/app/util"
	"github.com/AlexBond702/order-service/internal/pkg/http/httph"
	"github.com/AlexBond702/order-service/internal/pkg/http/mzerolog"
)

// NewHTTP создаёт HTTP процессор со служебными маршрутами (health и пр.).
func NewHttpAdmin(
	otelServiceName string,
	hHealth rhandler.Health,
	cfg section.ProcessorWebServer,
) *Processor {
	router := gin.New()
	if otelServiceName != "" {
		router.Use(otelgin.Middleware(
			otelServiceName,
			otelgin.WithFilter(func(r *http.Request) bool {
				return !util.IsFilteredHttpRoute(r)
			}),
		))
	}

	router.Use(
		httph.NewErrorMiddleware(),
		mzerolog.NewMiddleware(
			mzerolog.WithSkipper(util.IsFilteredHttpRoute)),
		makeErrorMiddleware(),
	)

	GenericRegHealthCheck(router, hHealth)
	GenericRegPprof(router)
	GenericRegMetrics(router)

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
