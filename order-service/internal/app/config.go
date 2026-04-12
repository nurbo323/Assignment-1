package app

import (
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    AppPort          string
    DBDSN            string
    PaymentGRPCAddr  string
    TrackingGRPCPort string
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

    return Config{
        AppPort:          port,
        DBDSN:            dsn,
        PaymentGRPCAddr:  paymentAddr,
        TrackingGRPCPort: trackingPort,
    }
}
