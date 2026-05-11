package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

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

	conn, err := amqp.Dial(cfg.RabbitURL)
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
	worker := worker.New(conn, cfg.QueueName, store, sender, cfg.RetryMaxAttempts, cfg.RetryBaseDelay, cfg.RetryMaxDelay)

	if err := worker.Run(ctx); err != nil {
		log.Fatalf("notification worker stopped: %v", err)
	}
}
