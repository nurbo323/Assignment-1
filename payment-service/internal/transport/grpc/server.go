package grpc

import (
	"context"
	"log"
	"net"

	paymentpb "github.com/nurbo323/generated-contracts/gen/payment"
	"google.golang.org/grpc"

	"payment-service/internal/usecase"
)

type Server struct {
	paymentpb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewServer(uc *usecase.PaymentUseCase) *Server {
	return &Server{uc: uc}
}

func (s *Server) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
	log.Printf("ProcessPayment called: order_id=%s amount=%d", req.OrderId, req.Amount)

	payment, err := s.uc.CreatePayment(ctx, usecase.CreatePaymentInput{
		OrderID: req.OrderId,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &paymentpb.PaymentResponse{
		PaymentId:     payment.ID,
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        string(payment.Status),
	}, nil
}

func (s *Server) GetPaymentStats(ctx context.Context, _ *paymentpb.GetPaymentStatsRequest) (*paymentpb.PaymentStats, error) {
	stats, err := s.uc.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &paymentpb.PaymentStats{
		TotalCount:      stats.TotalCount,
		AuthorizedCount: stats.AuthorizedCount,
		DeclinedCount:   stats.DeclinedCount,
		TotalAmount:     stats.TotalAmount,
	}, nil
}

func RunGRPCServer(uc *usecase.PaymentUseCase, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(grpcServer, NewServer(uc))

	log.Printf("gRPC payment-service listening on :%s", port)
	return grpcServer.Serve(lis)
}
