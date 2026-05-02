package usecase

import (
	"context"
	"encoding/json"
	"payment-service/internal/domain"

	"github.com/google/uuid"
)

type EventPublisher interface {
	PublishPaymentCompleted(ctx context.Context, payload []byte) error
}

type PaymentUseCase struct {
	repo      PaymentRepository
	publisher EventPublisher
}

type CreatePaymentInput struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

func NewPaymentUseCase(repo PaymentRepository, pub EventPublisher) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, publisher: pub}
}

func (u *PaymentUseCase) CreatePayment(ctx context.Context, input CreatePaymentInput) (domain.Payment, error) {
	if existing, err := u.repo.GetByOrderID(ctx, input.OrderID); err == nil {
		return existing, nil
	}

	payment := domain.Payment{ID: uuid.NewString(), OrderID: input.OrderID, TransactionID: uuid.NewString(), Amount: input.Amount, Status: domain.PaymentStatusAuthorized}
	if err := payment.Validate(); err != nil {
		return domain.Payment{}, err
	}
	if payment.Amount > 100000 {
		payment.Status = domain.PaymentStatusDeclined
	}

	if err := u.repo.Create(ctx, payment); err != nil {
		if existing, getErr := u.repo.GetByOrderID(ctx, input.OrderID); getErr == nil {
			return existing, nil
		}
		return domain.Payment{}, err
	}

	// Publish event after successful DB write. Use a simple JSON payload.
	if u.publisher != nil {
		evt := map[string]any{
			"order_id":       payment.OrderID,
			"amount":         payment.Amount,
			"customer_email": "user@example.com",
			"status":         string(payment.Status),
			"payment_id":     payment.ID,
		}
		if b, err := json.Marshal(evt); err == nil {
			_ = u.publisher.PublishPaymentCompleted(ctx, b)
		}
	}

	return payment, nil
}

func (u *PaymentUseCase) GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	return u.repo.GetByOrderID(ctx, orderID)
}

func (u *PaymentUseCase) GetStats(ctx context.Context) (domain.PaymentStats, error) {
	return u.repo.GetStats(ctx)
}
