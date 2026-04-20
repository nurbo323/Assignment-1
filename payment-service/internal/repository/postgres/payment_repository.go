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

func (r *PaymentRepository) GetStats(ctx context.Context) (domain.PaymentStats, error) {
	var stats domain.PaymentStats

	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) AS total_count,
			COALESCE(SUM(CASE WHEN status = 'Authorized' THEN 1 ELSE 0 END), 0) AS authorized_count,
			COALESCE(SUM(CASE WHEN status = 'Declined' THEN 1 ELSE 0 END), 0) AS declined_count,
			COALESCE(SUM(amount), 0) AS total_amount
		FROM payments
	`).Scan(
		&stats.TotalCount,
		&stats.AuthorizedCount,
		&stats.DeclinedCount,
		&stats.TotalAmount,
	)
	if err != nil {
		return domain.PaymentStats{}, err
	}

	return stats, nil
}
