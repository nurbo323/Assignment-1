package postgres

import (
	"context"
	"payment-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct{ db *pgxpool.Pool }

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository { return &PaymentRepository{db: db} }

func (r *PaymentRepository) Create(ctx context.Context, payment domain.Payment) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO payments (id, order_id, transaction_id, amount, status)
        VALUES ($1, $2, $3, $4, $5)
    `, payment.ID, payment.OrderID, payment.TransactionID, payment.Amount, payment.Status)
	return err
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	var payment domain.Payment
	err := r.db.QueryRow(ctx, `
        SELECT id, order_id, transaction_id, amount, status
        FROM payments
        WHERE order_id = $1
    `, orderID).Scan(&payment.ID, &payment.OrderID, &payment.TransactionID, &payment.Amount, &payment.Status)
	if err != nil {
		return domain.Payment{}, domain.ErrPaymentNotFound
	}
	return payment, nil
}
