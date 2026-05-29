package entity

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrOrderDuplicate       = errors.New("order already exist")
	ErrProductNotFound      = errors.New("product not found")
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrInvalidDeliveryPrice = errors.New("invalid delivery price")
	ErrInvalidID            = errors.New("invalid id")
)
