package main

import (
    "context"
    "log"

    "order-service/internal/app"
    repopg "order-service/internal/repository/postgres"
    grpctransport "order-service/internal/transport/grpc"
    httptransport "order-service/internal/transport/http"
    "order-service/internal/usecase"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    cfg := app.LoadConfig()

    ctx := context.Background()
    db, err := pgxpool.New(ctx, cfg.DBDSN)
    if err != nil {
        log.Fatalf("failed to connect to db: %v", err)
    }
    defer db.Close()

    orderRepo := repopg.NewOrderRepository(db)

    paymentClient, err := app.NewPaymentClient(cfg.PaymentGRPCAddr)
    if err != nil {
        log.Fatalf("failed to create payment grpc client: %v", err)
    }
    defer paymentClient.Close()

    broker := grpctransport.NewBroker()

    orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient, broker)

    go func() {
        if err := grpctransport.RunGRPCServer(broker, cfg.TrackingGRPCPort); err != nil {
            log.Fatalf("tracking grpc error: %v", err)
        }
    }()

    handler := httptransport.NewOrderHandler(orderUC)
    router := gin.Default()

    handler.RegisterRoutes(router)

    log.Printf("order-service listening on :%s", cfg.AppPort)
    if err := router.Run(":" + cfg.AppPort); err != nil {
        log.Fatalf("server error: %v", err)
    }
}
