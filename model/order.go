package model

import (
	"time"

	"github.com/google/uuid"
)

// Order models

type Order struct {
	OrderID     uint64     `json:"orderID"`
	CustomerID  uuid.UUID  `json:"customerID"`
	LineItems   []LineItem `json:"lineItems"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	ShippedAt   *time.Time `json:"shippedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type LineItem struct {
	ItemID   uuid.UUID `json:"itemID"`
	Quantity uint      `json:"quantity"`
	Price    uint      `json:"price"`
}

// Pagination models

type FindAllPage struct {
	Size   uint64 `json:"size,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
}

type FindAllResult struct {
	Orders []Order `json:"orders,omitempty"`
	Cursor uint64  `json:"cursor,omitempty"`
}
