package postgres

import (
	"context"
	"order-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct{ db *pgxpool.Pool }

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository { return &OrderRepository{db: db} }

func (r *OrderRepository) Create(ctx context.Context, order domain.Order) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, order.ID, order.CustomerID, order.ItemName, order.Amount, order.Status, order.CreatedAt)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (domain.Order, error) {
	var order domain.Order
	err := r.db.QueryRow(ctx, `
        SELECT id, customer_id, item_name, amount, status, created_at
        FROM orders
        WHERE id = $1
    `, id).Scan(&order.ID, &order.CustomerID, &order.ItemName, &order.Amount, &order.Status, &order.CreatedAt)
	if err != nil {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return order, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	tag, err := r.db.Exec(ctx, `UPDATE orders SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) GetIdempotency(ctx context.Context, key string) (domain.IdempotencyRecord, error) {
	var rec domain.IdempotencyRecord
	err := r.db.QueryRow(ctx, `
        SELECT idempotency_key, request_hash, order_id, created_at
        FROM order_idempotency
        WHERE idempotency_key = $1
    `, key).Scan(&rec.Key, &rec.RequestHash, &rec.OrderID, &rec.CreatedAt)
	if err != nil {
		return domain.IdempotencyRecord{}, domain.ErrOrderNotFound
	}
	return rec, nil
}

func (r *OrderRepository) SaveIdempotency(ctx context.Context, rec domain.IdempotencyRecord) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO order_idempotency (idempotency_key, request_hash, order_id, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (idempotency_key) DO NOTHING
    `, rec.Key, rec.RequestHash, rec.OrderID, rec.CreatedAt)
	if err != nil {
		return err
	}
	existing, err := r.GetIdempotency(ctx, rec.Key)
	if err != nil {
		return err
	}
	if existing.RequestHash != rec.RequestHash || existing.OrderID != rec.OrderID {
		return domain.ErrConflictIdempotency
	}
	return nil
}

func (r *OrderRepository) GetStats(ctx context.Context) (domain.OrderStats, error) {
	query := `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN status = 'Pending' THEN 1 ELSE 0 END), 0) AS pending,
			COALESCE(SUM(CASE WHEN status = 'Paid' THEN 1 ELSE 0 END), 0) AS paid,
			COALESCE(SUM(CASE WHEN status = 'Failed' THEN 1 ELSE 0 END), 0) AS failed,
			COALESCE(SUM(CASE WHEN status = 'Cancelled' THEN 1 ELSE 0 END), 0) AS cancelled
		FROM orders
	`

	var stats domain.OrderStats

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.Total,
		&stats.Pending,
		&stats.Paid,
		&stats.Failed,
		&stats.Cancelled,
	)
	if err != nil {
		return domain.OrderStats{}, err
	}

	return stats, nil
}
