package module

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/AlexBond702/order-service/internal/app/entity"
)

type Order interface {
	Create(ctx context.Context, userGUID uuid.UUID, deliveryPrice float64, currency string, items []entity.OrderItem) (entity.ResponseOrderCreate, error)
	Get(ctx context.Context, id int64) (entity.Order, error)
	Update(ctx context.Context, id int64, status string) (entity.ResponseOrderUpdate, error)
	Delete(ctx context.Context, id int64) error
}
type Metered interface {
	ProvideMetrics(fact promauto.Factory) []entity.MetricObservation
}
