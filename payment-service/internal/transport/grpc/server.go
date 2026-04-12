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

func RunGRPCServer(uc *usecase.PaymentUseCase) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(grpcServer, NewServer(uc))

	log.Println("gRPC payment-service listening on :50051")
	return grpcServer.Serve(lis)
}
