package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"order-service/internal/app"
	repopg "order-service/internal/repository/postgres"
	rediscache "order-service/internal/repository/redis"
	grpctransport "order-service/internal/transport/grpc"
	httptransport "order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := app.LoadConfig()

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := runMigrations(ctx, db, "/migrations/001_init.sql"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	orderRepo := repopg.NewOrderRepository(db)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()
	orderCache := rediscache.NewOrderCache(redisClient, cfg.CacheTTL)

	paymentClient, err := app.NewPaymentClient(cfg.PaymentGRPCAddr)
	if err != nil {
		log.Fatalf("failed to create payment grpc client: %v", err)
	}
	defer paymentClient.Close()

	broker := grpctransport.NewBroker()

	orderUC := usecase.NewOrderUseCase(orderRepo, orderCache, paymentClient, broker)

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

func runMigrations(ctx context.Context, db *pgxpool.Pool, path string) error {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}

	if _, err := db.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("exec migration %s: %w", path, err)
	}

	log.Printf("applied migration: %s", path)
	return nil
}
