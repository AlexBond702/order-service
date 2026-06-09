package morder

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AlexBond702/order-service/internal/app/client"
	"github.com/AlexBond702/order-service/internal/app/entity"
	rmodule "github.com/AlexBond702/order-service/internal/app/module"
	"github.com/AlexBond702/order-service/internal/app/repository"
)

type module struct {
	repoOrder repository.Order
	client    *client.CatalogClient
}

func NewModule(repoOrder repository.Order, client *client.CatalogClient) rmodule.Order {
	return &module{
		repoOrder: repoOrder,
		client:    client,
	}
}

func (m *module) Create(ctx context.Context, userGUID uuid.UUID, deliveryPrice float64, currency string, items []entity.OrderItem) (entity.ResponseOrderCreate, error) {
	if err := m.validateProductsWithCatalog(ctx, items); err != nil {
		return entity.ResponseOrderCreate{}, err
	}

	var CartPrice float64
	for _, item := range items {
		CartPrice += item.UnitPrice
	}
	CartPrice += deliveryPrice

	order := entity.Order{
		UserGuid:      userGUID,
		Status:        "pending",
		DeliveryPrice: deliveryPrice,
		Currency:      currency,
		Items:         items,
		TotalPrice:    CartPrice,
	}

	var createdOrder entity.Order
	err := m.repoOrder.OpenTx(ctx, func(c context.Context) error {
		var txErr error
		createdOrder, txErr = m.repoOrder.Create(c, order)
		return txErr
	})
	if err != nil {
		return entity.ResponseOrderCreate{}, fmt.Errorf("failed to created order: %w", err)
	}
	return m.convertToResponseCreate(createdOrder), nil
}

func (m *module) Get(ctx context.Context, id int64) (entity.Order, error) {
	order := entity.Order{
		ID: id,
	}
	var getOrder entity.Order
	err := m.repoOrder.OpenTx(ctx, func(c context.Context) error {
		var txErr error
		getOrder, txErr = m.repoOrder.Get(c, order.ID)
		return txErr
	})
	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to get order: %w", err)
	}
	return getOrder, nil
}

func (m *module) Update(ctx context.Context, id int64, newStatus string) (entity.ResponseOrderUpdate, error) {
	switch newStatus {
	case entity.OrderStatusCancelled, entity.OrderStatusPending, entity.OrderStatusDelivered, entity.OrderStatusShipped:
	default:
		return entity.ResponseOrderUpdate{}, fmt.Errorf("invalid status: %s", newStatus)
	}
	existingOrder, err := m.repoOrder.Get(ctx, id)
	if err != nil {
		return entity.ResponseOrderUpdate{}, fmt.Errorf("failed to get order: %w", err)
	}
	if existingOrder.Status == entity.OrderStatusPending && newStatus == entity.OrderStatusShipped {
		if err := m.validatePriceBeforeShipping(ctx, existingOrder); err != nil {
			return entity.ResponseOrderUpdate{}, err
		}
	}
	order := entity.Order{
		ID:     id,
		Status: newStatus,
	}
	var updateOrder entity.Order
	err = m.repoOrder.OpenTx(ctx, func(c context.Context) error {
		var txErr error
		updateOrder, txErr = m.repoOrder.Update(c, order)
		return txErr
	})
	if err != nil {
		return entity.ResponseOrderUpdate{}, fmt.Errorf("failed to updated order: %w", err)
	}
	return m.convertToResponseUpdate(updateOrder), nil
}

func (m *module) Delete(ctx context.Context, id int64) error {
	err := m.repoOrder.OpenTx(ctx, func(c context.Context) error {
		txErr := m.repoOrder.Delete(c, id)
		return txErr
	})
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}
	return nil
}

func (m *module) convertToResponseCreate(order entity.Order) entity.ResponseOrderCreate {
	response := entity.ResponseOrderCreate{
		ID:            order.ID,
		UserGUID:      order.UserGuid,
		DeliveryPrice: order.DeliveryPrice,
		Status:        order.Status,
		Currency:      order.Currency,
		CreatedAt:     order.CreatedAt.Format(time.RFC3339),
		Items:         make([]entity.ResponseOrderItem, 0, len(order.Items)),
	}
	for _, item := range order.Items {
		response.Items = append(response.Items, entity.ResponseOrderItem{
			ID:          item.ID,
			ProductGuid: item.ProductGUID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}
	return response
}

func (m *module) convertToResponseUpdate(order entity.Order) entity.ResponseOrderUpdate {
	response := entity.ResponseOrderUpdate{
		UserGUID:      order.UserGuid,
		DeliveryPrice: order.DeliveryPrice,
		Status:        order.Status,
		CreatedAt:     order.CreatedAt.Format(time.RFC3339),
		Items:         make([]entity.ResponseOrderItem, 0, len(order.Items)),
	}
	for _, item := range order.Items {
		response.Items = append(response.Items, entity.ResponseOrderItem{
			ID:          item.ID,
			ProductGuid: item.ProductGUID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}
	return response
}

func (m *module) validateProductsWithCatalog(ctx context.Context, items []entity.OrderItem) error {
	if items == nil {
		fmt.Printf("must not empty: %v", items)
	}
	productGuids := make([]uuid.UUID, len(items))
	for i, item := range items {
		productGuids[i] = item.ProductGUID
	}
	productsCatalog, err := m.client.GetProducts(ctx, productGuids)
	if err != nil {
		return fmt.Errorf("failed to validate products with catalog service: %w", err)
	}
	for _, item := range items {
		productCatalog := productsCatalog[item.ProductGUID.String()]
		if productCatalog.Price != item.UnitPrice {
			return entity.ErrProductPriceMismatch
		}
	}
	return nil
}

func (m *module) validatePriceBeforeShipping(ctx context.Context, order entity.Order) error {
	productGuids := make([]uuid.UUID, len(order.Items))
	for _, item := range order.Items {
		productGuids = append(productGuids, item.ProductGUID)
	}
	catalogProducts, err := m.client.GetProducts(ctx, productGuids)
	if err != nil {
		return fmt.Errorf("failed to check product prices: %w", err)
	}
	for _, item := range order.Items {
		catalogProduct := catalogProducts[item.ProductGUID.String()]
		if item.UnitPrice != catalogProduct.Price {
			return fmt.Errorf("cannot ship order: price for product %s has changed from %.2f to %.2f",
				item.ProductGUID, item.UnitPrice, catalogProduct.Price)
		}
	}
	return nil
}
