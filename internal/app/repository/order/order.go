package porder

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"

	"github.com/AlexBond702/order-service/internal/app/entity"
	"github.com/AlexBond702/order-service/internal/app/repository"
	rcpostgres "github.com/AlexBond702/order-service/internal/app/repository/conn/postgres"
	"github.com/AlexBond702/order-service/internal/app/repository/transaction"
)

type repoPg struct {
	db *gorm.DB
	*transaction.Repository
}

func NewOrderRepo(_ context.Context, client *rcpostgres.Client) (repository.Order, error) {
	txRepo := transaction.NewTxRepository(client.DB())
	return &repoPg{
		db:         client.DB(),
		Repository: txRepo,
	}, nil
}

func (r *repoPg) Create(ctx context.Context, order entity.Order) (entity.Order, error) {
	db := rcpostgres.GetTxFromContext(ctx, r.db)
	err := db.WithContext(ctx).
		Create(&order).Error
	if err != nil {
		return entity.Order{}, err
	}
	return order, nil
}

func (r *repoPg) Get(ctx context.Context, id int64) (entity.Order, error) {
	var order entity.Order
	db := rcpostgres.GetTxFromContext(ctx, r.db)

	result := db.WithContext(ctx).
		Preload("Items").
		Where("id=?", id).
		First(&order)

	if result.Error != nil {
		return entity.Order{}, result.Error
	}
	if result.RowsAffected == 0 {
		return entity.Order{}, entity.ErrNotFound
	}
	return order, nil
}

func (r *repoPg) Update(ctx context.Context, order entity.Order) (entity.Order, error) {
	db := rcpostgres.GetTxFromContext(ctx, r.db)
	result := db.WithContext(ctx).
		Model(&entity.Order{}).
		Where("id=?", order.ID).
		Update("status", order.Status)
	if result.Error != nil {
		return entity.Order{}, result.Error
	}
	if result.RowsAffected == 0 {
		return entity.Order{}, entity.ErrNotFound
	}
	return order, nil
}

func (r *repoPg) Delete(ctx context.Context, id int64) error {
	db := rcpostgres.GetTxFromContext(ctx, r.db)
	result := db.WithContext(ctx).
		Where("id=?", id).
		Delete(&entity.Order{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *repoPg) ApplyUpdateOrder(ctx context.Context, orderGUID uuid.UUID, deliveryPrice int64) (bool, error) {
	db := rcpostgres.GetTxFromContext(ctx, r.db)
	result := db.WithContext(ctx).
		Model(&entity.Order{}).
		Where("guid=?", orderGUID).
		Where("status=?", entity.OrderStatusPending).
		Updates(map[string]any{
			"delivery_price": deliveryPrice,
			"status":         entity.OrderStatusDeliveryCalculated,
		})

	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}
