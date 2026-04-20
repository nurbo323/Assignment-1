package usecase

import (
	"context"
	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error)
	GetStats(ctx context.Context) (domain.PaymentStats, error)
}
