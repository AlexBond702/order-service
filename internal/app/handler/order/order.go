package horder

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/AlexBond702/order-service/internal/app/entity"
	handler "github.com/AlexBond702/order-service/internal/app/handler"
	"github.com/AlexBond702/order-service/internal/app/module"
	"github.com/AlexBond702/order-service/internal/pkg/http/httph"
)

type handlerOrder struct {
	moduleOrder module.Order
}

func NewHandler(moduleOrder module.Order) handler.Order {
	return &handlerOrder{
		moduleOrder: moduleOrder,
	}
}

func (h *handlerOrder) Create(c *gin.Context) {
	var req entity.RequestOrderCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}
	items := make([]entity.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, entity.OrderItem{
			ProductGUID: item.ProductGuid,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	response, err := h.moduleOrder.Create(
		c.Request.Context(),
		req.UserGUID,
		req.Currency,
		items)
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}
	httph.SendEncoded(c.Writer, c.Request, http.StatusCreated, response)
}

func (h *handlerOrder) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}
	response, err := h.moduleOrder.Get(c.Request.Context(), id)
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	httph.SendEncoded(c.Writer, c.Request, http.StatusOK, response)
}

func (h *handlerOrder) Update(c *gin.Context) {
	var req entity.RequestOrderUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}
	response, err := h.moduleOrder.Update(c.Request.Context(), req.ID, req.Status)
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}
	httph.SendEncoded(c.Writer, c.Request, http.StatusOK, response)
}

func (h *handlerOrder) Delete(c *gin.Context) {
	var req entity.Order
	if err := c.ShouldBindJSON(&req); err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}
	err := h.moduleOrder.Delete(c.Request.Context(), req.ID)
	if err != nil {
		httph.ErrorApply(c.Request, err)
	}
	httph.SendEncoded(c.Writer, c.Request, http.StatusNoContent, nil)
}
