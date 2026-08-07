package ehandler

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog/log"

	"github.com/AlexBond702/order-service/internal/app/entity"
	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
	service "github.com/AlexBond702/order-service/internal/app/module"
	"github.com/AlexBond702/order-service/internal/pkg/broker"
)

type handler struct {
	srv service.UpdateDelivery
}

func NewHandlerOrderDelivery(srv service.UpdateDelivery) rhandler.OrderDelivery {
	return &handler{
		srv: srv,
	}
}

func (h *handler) CallbackOrderDelivery(ctx context.Context,
	ev *entity.EventOrderDeliveryCalculated,
	_ []broker.Header,
) error {
	log.Info().Ctx(ctx).EmbedObject(ev).Msg("Получено событие order.delivery.calculated")
	if ev.OrderGUID == "" {
		return broker.NotCriticalError(fmt.Errorf("order.delivery.calculated: empty order_guid"))
	}
	orderGuid, err := uuid.FromString(ev.OrderGUID)
	if err != nil {
		return broker.NotCriticalError(fmt.Errorf(
			"order.delivery.calculated: failed parse %s: %w",
			ev.OrderGUID, err))
	}

	apply, err := h.srv.ApplyUpdateDelivery(ctx, orderGuid, ev.DeliveryPrice)
	if err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}
	if !apply {
		log.Warn().
			Str("order_guid", ev.OrderGUID).
			Msg("delivery skipped: order is not pending")
		return nil
	}
	log.Info().
		Str("order_guid", ev.OrderGUID).
		Int64("delivery_price", ev.DeliveryPrice).
		Msg("delivery applied")

	return nil
}
