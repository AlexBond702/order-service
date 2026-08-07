package emonitor

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/AlexBond702/order-service/internal/app/entity"
	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
	"github.com/AlexBond702/order-service/internal/app/processor"
	"github.com/AlexBond702/order-service/internal/pkg/broker"
)

type orderDeliveryCalculateProc struct {
	h   rhandler.OrderDelivery
	bus broker.Bus[entity.EventOrderDeliveryCalculated]
}

func NewProc(h rhandler.OrderDelivery,
	bus broker.Bus[entity.EventOrderDeliveryCalculated],
) processor.Processor {
	return &orderDeliveryCalculateProc{
		h:   h,
		bus: bus,
	}
}

func (o *orderDeliveryCalculateProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	if err := o.bus.Subscribe(ctx, wg, o.h.CallbackOrderDelivery); err != nil {
		log.Fatal().Err(err).Msg("order.delivery.calculated: failed to start subscribe")
	}
	log.Info().Str("queue_name", o.bus.QueueName()).Msg("subscribe requested")
}
