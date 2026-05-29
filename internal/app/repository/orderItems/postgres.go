package orderItems

import (
	"context"

	"gorm.io/gorm"

	"github.com/AlexBond702/order-service/internal/app/entity"
	repository "github.com/AlexBond702/order-service/internal/app/repository/conn/postgres"
	"github.com/AlexBond702/order-service/internal/app/repository/transaction"
)

type OrderItems struct {
	db *gorm.DB
	transaction.Repository
}

func (r *OrderItems) Update(ctx context.Context, order entity.Order) error {
	db := repository.GetTxFromContext(ctx, r.db)
	result := db.WithContext(ctx).
		Where("id=?", order.ID).
		Updates(&order)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *OrderItems) Delete(ctx context.Context, id int64) error {
	db := repository.GetTxFromContext(ctx, r.db)
	result := db.WithContext(ctx).
		Where("id=?", id).
		Delete(id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}
