package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"payment-service/internal/app"
	"payment-service/internal/messaging"
	repopg "payment-service/internal/repository/postgres"
	grpctransport "payment-service/internal/transport/grpc"
	httptransport "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
	"time"

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

	if err := runMigrations(ctx, db, "/migrations/001_init.sql"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	repo := repopg.NewPaymentRepository(db)

	// initialize rabbitmq publisher
	rabbitURL := "amqp://guest:guest@rabbitmq:5672/"
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		rabbitURL = v
	}
	pub, err := newPublisherWithRetry(rabbitURL, "payment.completed", 5, 2*time.Second)
	if err != nil {
		log.Printf("failed to init rabbitmq publisher after retries: %v", err)
		pub = nil
	}

	uc := usecase.NewPaymentUseCase(repo, pub)
	handler := httptransport.NewPaymentHandler(uc)

	go func() {
		if err := grpctransport.RunGRPCServer(uc, cfg.GRPCPort); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	router := gin.Default()
	handler.RegisterRoutes(router)

	log.Printf("payment-service listening on :%s", cfg.AppPort)
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

func newPublisherWithRetry(url string, queueName string, attempts int, delay time.Duration) (usecase.EventPublisher, error) {
	if attempts <= 0 {
		attempts = 1
	}
	if delay <= 0 {
		delay = 2 * time.Second
	}

	var lastErr error
	currentDelay := delay
	for attempt := 1; attempt <= attempts; attempt++ {
		pub, err := messaging.NewRabbitPublisher(url, queueName)
		if err == nil {
			return pub, nil
		}
		lastErr = err
		log.Printf("rabbitmq publisher init failed attempt=%d err=%v", attempt, err)
		time.Sleep(currentDelay)
		if currentDelay < 10*time.Second {
			currentDelay *= 2
			if currentDelay > 10*time.Second {
				currentDelay = 10 * time.Second
			}
		}
	}

	return nil, lastErr
}
