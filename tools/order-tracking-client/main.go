package main

import (
    "context"
    "flag"
    "log"

    orderpb "github.com/nurbo323/generated-contracts/gen/order"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    addr := flag.String("addr", "localhost:50052", "order tracking grpc address")
    orderID := flag.String("order-id", "", "order id to subscribe")
    flag.Parse()

    if *orderID == "" {
        log.Fatal("order-id is required")
    }

    conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()

    client := orderpb.NewOrderTrackingServiceClient(conn)

    stream, err := client.SubscribeToOrderUpdates(context.Background(), &orderpb.OrderRequest{
        OrderId: *orderID,
    })
    if err != nil {
        log.Fatalf("failed to subscribe: %v", err)
    }

    for {
        msg, err := stream.Recv()
        if err != nil {
            log.Fatalf("stream error: %v", err)
        }

        log.Printf("STATUS UPDATE: %s", msg.Status)
    }
}
