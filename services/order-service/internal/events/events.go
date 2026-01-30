package events

import (
    "time"
    "github.com/google/uuid"
)

type OrderItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

type OrderCreatedEvent struct {
	OrderID   uuid.UUID   `json:"order_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Items     []OrderItem `json:"items"`
	Timestamp time.Time   `json:"timestamp"`
}

type OrderCancelledEvent struct {
	OrderID   uuid.UUID   `json:"order_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Items     []OrderItem `json:"items"`
	Timestamp time.Time   `json:"timestamp"`
}
