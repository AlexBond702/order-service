package eorder

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"github.com/AlexBond702/order-service/internal/app/module"
	"github.com/AlexBond702/order-service/internal/app/repository"
)

type srv struct {
	repo repository.Order
}

func NewServiceUpdateDelivery(repo repository.Order) module.UpdateDelivery {
	return &srv{
		repo: repo,
	}
}

func (s *srv) ApplyUpdateDelivery(ctx context.Context,
	orderGUID uuid.UUID,
	deliveryPrice int64,
) (bool, error) {
	return s.repo.ApplyUpdateOrder(ctx, orderGUID, deliveryPrice)
}
