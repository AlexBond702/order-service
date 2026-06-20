package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
)

func GenericRegHealthCheck(r *gin.Engine, h rhandler.Health) {
	regRoute(r, http.MethodGet, "/health", h.LastCheck, "api.generic.health_check")
}

func GenericRegMetrics(r *gin.Engine) {
	regRoute(r, http.MethodGet, "/metrics", gin.WrapH(promhttp.Handler()), "api.generic.metrics")
}
