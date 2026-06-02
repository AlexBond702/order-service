package entity

import (
	"time"

	"github.com/google/uuid"
)

// //////////////////////////////////////////////////////////////////////////////
// /// DATABASE MODELS //////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////////////////////////
type Order struct {
	ID            int64       `gorm:"autoincrement;not null;unique"`
	GUID          uuid.UUID   `gorm:"primaryKey;default:gen_random_uuid()"`
	UserGuid      uuid.UUID   `gorm:"column:user_guid;default:null"`
	TotalPrice    float64     `gorm:"column:total_price;default:0"`
	DeliveryPrice float64     `gorm:"column:delivery_price;default:0"`
	Currency      string      `gorm:"type:text;not null"`
	Status        string      `gorm:"type:text;not null"`
	CreatedAt     time.Time   `gorm:"column:created_at;default:NOW();not null"`
	UpdatedAt     time.Time   `gorm:"column:updated_at;default:NOW();not null"`
	Items         []OrderItem `gorm:"foreignKey:OrderGUID;references:GUID;constraint:OnDelete:CASCADE"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID          int64     `gorm:"autoincrement;not null;unique"`
	GUID        uuid.UUID `gorm:"primaryKey;default:gen_random_uuid()"`
	OrderGUID   uuid.UUID `gorm:"column:order_guid;not null"`
	ProductGUID uuid.UUID `gorm:"column:product_guid;not null"`
	Quantity    int       `gorm:"not null"`
	UnitPrice   float64   `gorm:"column:unit_price;no null"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

////////////////////////////////////////////////////////////////////////////////
///// HTTP REQUEST & RESPONSES /////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

type ResponseOrderCreate struct {
	ID            int64               `json:"id"`
	UserGUID      uuid.UUID           `json:"user_guid"`
	DeliveryPrice float64             `json:"delivery_price"`
	Status        string              `json:"status"`
	Currency      string              `json:"currency"`
	CreatedAt     string              `json:"created_at"`
	Items         []ResponseOrderItem `json:"items"`
}

type ResponseOrderUpdate struct {
	UserGUID      uuid.UUID           `json:"user_guid"`
	DeliveryPrice float64             `json:"delivery_price"`
	Status        string              `json:"status"`
	CreatedAt     string              `json:"created_at"`
	Items         []ResponseOrderItem `json:"items"`
}

type ResponseOrderItem struct {
	ID          int64     `json:"id"`
	ProductGuid uuid.UUID `json:"product_guid"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
}
type RequestOrderCreate struct {
	UserGUID      uuid.UUID          `json:"user_guid" binding:"required,uuid"`
	DeliveryPrice float64            `json:"delivery_price" binding:"required,numeric,ne=0,gt=0"`
	Currency      string             `json:"currency" binding:"omitempty,max=255"`
	Items         []RequestOrderItem `json:"items" binding:"required"`
}
type RequestOrderUpdate struct {
	ID     int64  `json:"id" binding:"required"`
	Status string `json:"status" binding:"required,min=2,max=255"`
}

type RequestOrderItem struct {
	ProductGuid uuid.UUID `json:"product_guid" binding:"required,uuid"`
	Quantity    int       `json:"quantity" binding:"required"`
	UnitPrice   float64   `json:"unit_price" binding:"required,numeric,ne=0,gt=0"`
}
