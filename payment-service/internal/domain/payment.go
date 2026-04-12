package domain

import "errors"

const (
	PaymentStatusAuthorized = "Authorized"
	PaymentStatusDeclined   = "Declined"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidAmount   = errors.New("amount must be greater than 0")
)

type Payment struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

func (p Payment) Validate() error {
	if p.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}
