package grpc

import (
    "io"
    "log"
    "net"

    orderpb "github.com/nurbo323/generated-contracts/gen/order"
    "google.golang.org/grpc"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
    orderpb.UnimplementedOrderTrackingServiceServer
    broker *Broker
}

func NewServer(broker *Broker) *Server {
    return &Server{broker: broker}
}

func (s *Server) SubscribeToOrderUpdates(
    req *orderpb.OrderRequest,
    stream orderpb.OrderTrackingService_SubscribeToOrderUpdatesServer,
) error {
    orderID := req.OrderId
    log.Printf("tracking subscriber connected: order_id=%s", orderID)

    ch, unsubscribe := s.broker.Subscribe(orderID)
    defer unsubscribe()

    for {
        select {
        case <-stream.Context().Done():
            log.Printf("tracking subscriber disconnected: order_id=%s", orderID)
            return nil

        case event, ok := <-ch:
            if !ok {
                return io.EOF
            }

            err := stream.Send(&orderpb.OrderStatusUpdate{
                OrderId:   event.OrderID,
                Status:    event.Status,
                UpdatedAt: timestamppb.New(event.UpdatedAt),
            })
            if err != nil {
                return err
            }
        }
    }
}

func RunGRPCServer(broker *Broker, port string) error {
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        return err
    }

    s := grpc.NewServer()
    orderpb.RegisterOrderTrackingServiceServer(s, NewServer(broker))

    log.Printf("order tracking gRPC listening on :%s", port)
    return s.Serve(lis)
}
