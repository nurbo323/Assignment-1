package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/app"
	"notification-service/internal/provider"
	"notification-service/internal/state"
	"notification-service/internal/worker"

	amqp "github.com/rabbitmq/amqp091-go"
	redislib "github.com/redis/go-redis/v9"
)

func main() {
	cfg := app.LoadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conn, err := dialRabbitWithRetry(ctx, cfg.RabbitURL, 5, 2*time.Second)
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	defer conn.Close()

	redisClient := redislib.NewClient(&redislib.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	sender, err := provider.NewSender(cfg)
	if err != nil {
		log.Fatalf("failed to initialize notification provider: %v", err)
	}

	store := state.NewStore(redisClient, cfg.ProcessingLockTTL, cfg.SentTTL, cfg.FailedTTL)
	worker := worker.New(conn, cfg.QueueName, store, sender, cfg.RetryMaxAttempts, cfg.RetryBaseDelay, cfg.RetryMaxDelay, cfg.WorkerConcurrency)

	if err := worker.Run(ctx); err != nil {
		log.Fatalf("notification worker stopped: %v", err)
	}
}

func dialRabbitWithRetry(ctx context.Context, url string, attempts int, delay time.Duration) (*amqp.Connection, error) {
	if attempts <= 0 {
		attempts = 1
	}
	if delay <= 0 {
		delay = 2 * time.Second
	}

	var lastErr error
	currentDelay := delay
	for attempt := 1; attempt <= attempts; attempt++ {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		log.Printf("rabbitmq connect failed attempt=%d err=%v", attempt, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(currentDelay):
		}
		if currentDelay < 10*time.Second {
			currentDelay *= 2
			if currentDelay > 10*time.Second {
				currentDelay = 10 * time.Second
			}
		}
	}

	return nil, lastErr
}
