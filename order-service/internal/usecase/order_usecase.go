package usecase

import (
    "context"
    "errors"
    "order-service/internal/domain"
    "time"

    "github.com/google/uuid"
)

type OrderUseCase struct {
    repo      OrderRepository
    payment   PaymentAuthorizer
    publisher OrderStatusPublisher
}

type CreateOrderInput struct {
    CustomerID string `json:"customer_id"`
    ItemName   string `json:"item_name"`
    Amount     int64  `json:"amount"`
}

func NewOrderUseCase(repo OrderRepository, payment PaymentAuthorizer, publisher OrderStatusPublisher) *OrderUseCase {
    return &OrderUseCase{
        repo:      repo,
        payment:   payment,
        publisher: publisher,
    }
}

func (u *OrderUseCase) CreateOrder(ctx context.Context, input CreateOrderInput) (domain.Order, error) {
    if input.CustomerID == "" || input.ItemName == "" || input.Amount <= 0 {
        return domain.Order{}, errors.New("invalid input")
    }

    order := domain.Order{
        ID:         uuid.NewString(),
        CustomerID: input.CustomerID,
        ItemName:   input.ItemName,
        Amount:     input.Amount,
        Status:     "Pending",
        CreatedAt:  time.Now(),
    }

    if err := u.repo.Create(ctx, order); err != nil {
        return domain.Order{}, err
    }

    res, err := u.payment.Authorize(ctx, order.ID, order.Amount)
    if err != nil {
        _ = u.repo.UpdateStatus(ctx, order.ID, "Failed")
        if u.publisher != nil {
            u.publisher.Publish(order.ID, "Failed")
        }
        order.Status = "Failed"
        return order, nil
    }

    if res.Status == "Authorized" || res.Status == "Paid" || res.Status == "SUCCESS" {
        _ = u.repo.UpdateStatus(ctx, order.ID, "Paid")
        if u.publisher != nil {
            u.publisher.Publish(order.ID, "Paid")
        }
        order.Status = "Paid"
    } else {
        _ = u.repo.UpdateStatus(ctx, order.ID, "Declined")
        if u.publisher != nil {
            u.publisher.Publish(order.ID, "Declined")
        }
        order.Status = "Declined"
    }

    return order, nil
}

func (u *OrderUseCase) GetOrder(ctx context.Context, id string) (domain.Order, error) {
    return u.repo.GetByID(ctx, id)
}

func (u *OrderUseCase) CancelOrder(ctx context.Context, id string) (domain.Order, error) {
    order, err := u.repo.GetByID(ctx, id)
    if err != nil {
        return domain.Order{}, err
    }

    if order.Status == "Cancelled" {
        return order, nil
    }

    if err := u.repo.UpdateStatus(ctx, id, "Cancelled"); err != nil {
        return domain.Order{}, err
    }

    if u.publisher != nil {
        u.publisher.Publish(id, "Cancelled")
    }

    order.Status = "Cancelled"
    return order, nil
}

func (u *OrderUseCase) GetStats(ctx context.Context) (domain.OrderStats, error) {
    return u.repo.GetStats(ctx)
}
