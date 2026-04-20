package app

import (
	"context"
	"time"

	paymentpb "github.com/nurbo323/generated-contracts/gen/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"order-service/internal/usecase"
)

type PaymentClient struct {
	client paymentpb.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewPaymentClient(addr string) (*PaymentClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &PaymentClient{
		client: paymentpb.NewPaymentServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *PaymentClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *PaymentClient) Authorize(ctx context.Context, orderID string, amount int64) (usecase.PaymentAuthorizationResult, error) {
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	res, err := c.client.ProcessPayment(callCtx, &paymentpb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		return usecase.PaymentAuthorizationResult{}, err
	}

	return usecase.PaymentAuthorizationResult{
		Status:        res.Status,
		TransactionID: res.TransactionId,
	}, nil
}

func (c *PaymentClient) GetPaymentStats(ctx context.Context) (usecase.PaymentStats, error) {
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	res, err := c.client.GetPaymentStats(callCtx, &paymentpb.GetPaymentStatsRequest{})
	if err != nil {
		return usecase.PaymentStats{}, err
	}

	return usecase.PaymentStats{
		TotalCount:      res.TotalCount,
		AuthorizedCount: res.AuthorizedCount,
		DeclinedCount:   res.DeclinedCount,
		TotalAmount:     res.TotalAmount,
	}, nil
}
