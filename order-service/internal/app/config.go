package app

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort          string
	DBDSN            string
	PaymentGRPCAddr  string
	TrackingGRPCPort string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	CacheTTL         time.Duration
}

func LoadConfig() Config {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:2006@localhost:5432/orders_db?sslmode=disable"
	}

	paymentAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if paymentAddr == "" {
		paymentAddr = "localhost:50051"
	}

	trackingPort := os.Getenv("TRACKING_GRPC_PORT")
	if trackingPort == "" {
		trackingPort = "50052"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if raw := os.Getenv("REDIS_DB"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			redisDB = parsed
		}
	}

	cacheTTL := 5 * time.Minute
	if raw := os.Getenv("CACHE_TTL"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			cacheTTL = parsed
		}
	}

	return Config{
		AppPort:          port,
		DBDSN:            dsn,
		PaymentGRPCAddr:  paymentAddr,
		TrackingGRPCPort: trackingPort,
		RedisAddr:        redisAddr,
		RedisPassword:    redisPassword,
		RedisDB:          redisDB,
		CacheTTL:         cacheTTL,
	}
}
