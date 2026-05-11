package app

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	RabbitURL            string
	QueueName            string
	ProviderMode         string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	ProcessingLockTTL    time.Duration
	SentTTL              time.Duration
	FailedTTL            time.Duration
	RetryMaxAttempts     int
	RetryBaseDelay       time.Duration
	RetryMaxDelay        time.Duration
	SimulatedLatency     time.Duration
	SimulatedFailureRate float64
	SMTPHost             string
	SMTPPort             string
	SMTPUsername         string
	SMTPPassword         string
	SMTPFrom             string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	return Config{
		RabbitURL:            envString("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		QueueName:            envString("NOTIFICATION_QUEUE", "payment.completed"),
		ProviderMode:         envString("PROVIDER_MODE", "SIMULATED"),
		RedisAddr:            envString("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        os.Getenv("REDIS_PASSWORD"),
		RedisDB:              envInt("REDIS_DB", 0),
		ProcessingLockTTL:    envDuration("PROCESSING_LOCK_TTL", 30*time.Second),
		SentTTL:              envDuration("NOTIFICATION_SENT_TTL", 24*time.Hour),
		FailedTTL:            envDuration("NOTIFICATION_FAILED_TTL", 1*time.Hour),
		RetryMaxAttempts:     envInt("RETRY_MAX_ATTEMPTS", 3),
		RetryBaseDelay:       envDuration("RETRY_BASE_DELAY", 2*time.Second),
		RetryMaxDelay:        envDuration("RETRY_MAX_DELAY", 8*time.Second),
		SimulatedLatency:     envDuration("SIMULATED_LATENCY", 500*time.Millisecond),
		SimulatedFailureRate: envFloat("SIMULATED_FAILURE_RATE", 0.2),
		SMTPHost:             os.Getenv("SMTP_HOST"),
		SMTPPort:             envString("SMTP_PORT", "587"),
		SMTPUsername:         os.Getenv("SMTP_USERNAME"),
		SMTPPassword:         os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:             os.Getenv("SMTP_FROM"),
	}
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return fallback
}
