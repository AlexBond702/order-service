package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
)

func GenericRegHealthCheck(r *gin.Engine, h rhandler.Health) {
	regRoute(r, http.MethodGet, "/health", h.LastCheck, "api.generic.health_check")
}
