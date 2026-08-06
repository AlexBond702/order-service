package repository

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"github.com/AlexBond702/order-service/internal/app/entity"
)

type Transactional interface {
	OpenTx(ctx context.Context, fn func(c context.Context) error) error
}

type Order interface {
	Transactional
	Create(ctx context.Context, order entity.Order) (entity.Order, error)
	Get(ctx context.Context, id int64) (entity.Order, error)
	Update(ctx context.Context, order entity.Order) (entity.Order, error)
	Delete(ctx context.Context, id int64) error
	ApplyUpdateOrder(ctx context.Context, orderGUID uuid.UUID, deliveryPrice int64) (bool, error)
}
