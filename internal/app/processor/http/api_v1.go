package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
)

func v1GenericRegOrder(r *gin.RouterGroup, h rhandler.Order) {
	regRoute(r, http.MethodPost, "/order", h.Create, "api.v1.create_order")
	regRoute(r, http.MethodGet, "/order/:id", h.Get, "api.v1.get_order")
	regRoute(r, http.MethodDelete, "/order/:id", h.Delete, "api.v1.delete_order")
	regRoute(r, http.MethodPatch, "/order/:id", h.Update, "api.v1.create_order")
}
