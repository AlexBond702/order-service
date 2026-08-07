package rhandler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/AlexBond702/order-service/internal/app/entity"
	"github.com/AlexBond702/order-service/internal/pkg/broker"
)

type Health interface {
	LastCheck(c *gin.Context)
}

type Order interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type OrderDelivery interface {
	CallbackOrderDelivery(ctx context.Context,
		ev *entity.EventOrderDeliveryCalculated,
		headers []broker.Header,
	) error
}
