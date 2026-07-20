package entity

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"

	"github.com/AlexBond702/order-service/internal/pkg/broker"
)

const (
	OrderStatusPending   = "pending"
	OrderStatusCancelled = "cancelled"
	OrderStatusDelivered = "delivered"
	OrderStatusShipped   = "shipped"
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
	TotalPrice    float64             `json:"total_price"`
	DeliveryPrice float64             `json:"delivery_price"`
	Status        string              `json:"status"`
	Currency      string              `json:"currency"`
	CreatedAt     string              `json:"created_at"`
	Items         []ResponseOrderItem `json:"items"`
}

type ResponseOrderUpdate struct {
	UserGUID      uuid.UUID           `json:"user_guid"`
	TotalPrice    float64             `json:"total_price"`
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
	UserGUID      uuid.UUID          `json:"user_guid" binding:"omitempty,uuid"`
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

////////////////////////////////////////////////////////////////////////////////
///// EVENT MODEL //////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

type EventOrderCreated struct {
	OrderGUID  string                  `json:"order_guid"`
	UserGUID   *string                 `json:"user_guid,omitempty"`
	Currency   string                  `json:"currency"`
	TotalPrice int64                   `json:"total_price"`
	Items      []EventOrderCreatedItem `json:"items"`
	CreatedAt  string                  `json:"created_at"`
}
type EventOrderCreatedItem struct {
	ProductGUID string `json:"product_guid"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
}

func (e EventOrderCreated) EventId() string {
	return e.OrderGUID
}

func (e EventOrderCreated) MarshalZerologObject(ev *zerolog.Event) {
	ev.Str("order_guid", e.OrderGUID).
		Str("currency", e.Currency).
		Int64("total_price", e.TotalPrice).
		Int("items_count", len(e.Items))
}

////////////////////////////////////////////////////////////////////////////////
///// EVENT AUXILIARIES ////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

const (
	BrokerHeaderKeyOrderEventType = "type"
	BrokerHeaderKeyOrderEventID   = "event-id"

	BrokerHeaderValueOrderCreated = "order.created"
)

func BrokerHeaderOrderCreatedType() broker.Header {
	return broker.Header{
		Key:   BrokerHeaderKeyOrderEventType,
		Value: BrokerHeaderValueOrderCreated,
	}
}

func BrokerHeaderOrderCreatedEventID() broker.Header {
	return broker.Header{
		Key:   BrokerHeaderKeyOrderEventID,
		Value: uuid.Must(uuid.NewV4()).String(),
	}
}
