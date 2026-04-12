package app

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort  string
	DBDSN    string
	GRPCPort string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8081"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:2006@localhost:5432/payments_db?sslmode=disable"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	return Config{
		AppPort:  port,
		DBDSN:    dsn,
		GRPCPort: grpcPort,
	}
}
