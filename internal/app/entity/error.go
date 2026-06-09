package entity

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrOrderDuplicate       = errors.New("order already exist")
	ErrProductNotFound      = errors.New("product not found")
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrInvalidDeliveryPrice = errors.New("invalid delivery price")
	ErrInvalidID            = errors.New("invalid id")

	ErrOrderNotFound       = errors.New("order not found")
	ErrInvalidOrderID      = errors.New("invalid order ID")
	ErrOrderCreationFailed = errors.New("failed to create order")

	ErrProductNotFoundInCatalog = errors.New("product not found in catalog")
	ErrProductPriceMismatch     = errors.New("product price mismatch")
	ErrProductPriceChanged      = errors.New("product price has changed")

	ErrCatalogServiceUnavailable = errors.New("catalog service unavailable")
)
