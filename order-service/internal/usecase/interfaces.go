package usecase

import (
    "context"
    "order-service/internal/domain"
)

type OrderRepository interface {
    Create(ctx context.Context, order domain.Order) error
    GetByID(ctx context.Context, id string) (domain.Order, error)
    UpdateStatus(ctx context.Context, id string, status string) error
    GetIdempotency(ctx context.Context, key string) (domain.IdempotencyRecord, error)
    SaveIdempotency(ctx context.Context, rec domain.IdempotencyRecord) error
    GetStats(ctx context.Context) (domain.OrderStats, error)
}

type PaymentAuthorizer interface {
    Authorize(ctx context.Context, orderID string, amount int64) (PaymentAuthorizationResult, error)
}

type OrderStatusPublisher interface {
    Publish(orderID string, status string)
}

type PaymentAuthorizationResult struct {
    Status        string `json:"status"`
    TransactionID string `json:"transaction_id"`
}

type OrderService interface {
    CreateOrder(ctx context.Context, input CreateOrderInput) (domain.Order, error)
    GetOrder(ctx context.Context, id string) (domain.Order, error)
    CancelOrder(ctx context.Context, id string) (domain.Order, error)
    GetStats(ctx context.Context) (domain.OrderStats, error)
}
