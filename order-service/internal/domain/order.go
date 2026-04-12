package domain

import (
	"errors"
	"time"
)

const (
	OrderStatusPending   = "Pending"
	OrderStatusPaid      = "Paid"
	OrderStatusFailed    = "Failed"
	OrderStatusCancelled = "Cancelled"
)

var (
	ErrInvalidAmount             = errors.New("amount must be greater than 0")
	ErrOrderNotFound             = errors.New("order not found")
	ErrOrderCannotBeCancelled    = errors.New("only pending orders can be cancelled")
	ErrPaymentServiceUnavailable = errors.New("payment service unavailable")
	ErrConflictIdempotency       = errors.New("idempotency key already used with a different payload")
)

type Order struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	ItemName   string    `json:"item_name"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type IdempotencyRecord struct {
	Key         string
	RequestHash string
	OrderID     string
	CreatedAt   time.Time
}

func (o Order) ValidateForCreate() error {
	if o.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}
