package porder

import (
	"context"

	"gorm.io/gorm"

	"github.com/AlexBond702/order-service/internal/app/entity"
	repository "github.com/AlexBond702/order-service/internal/app/repository/conn/postgres"
	"github.com/AlexBond702/order-service/internal/app/repository/transaction"
)

type OrderRepo struct {
	db *gorm.DB
	*transaction.Repository
}

func (r *OrderRepo) GetByID(ctx context.Context, id int64) (entity.Order, error) {
	///var order entity.Order
	///db := repository.GetTxFromContext(ctx, r.db)

	///err := db.WithContext(ctx)
	////
	///return order, util.ReplaceErr1(err, gorm.ErrRecordNotFound, entity.ErrNotFound)
	return entity.Order{}, nil
}

func (r *OrderRepo) Update(ctx context.Context, order entity.Order) error {
	db := repository.GetTxFromContext(ctx, r.db)
	result := db.WithContext(ctx).
		Model(&entity.Order{}).
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

func (r *OrderRepo) Delete(ctx context.Context, id int64) error {
	db := r.db.WithContext(ctx)
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
