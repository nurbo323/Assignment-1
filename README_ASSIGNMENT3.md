# Assignment 3 — Event-Driven Architecture (EDA) with Message Queues

## Overview

This assignment implements an asynchronous notification system using **RabbitMQ** as a message broker. The Payment Service publishes events after successful payment processing, and the Notification Service consumes them with full reliability guarantees:

- **At-least-once delivery**: Messages are never lost if a consumer crashes.
- **Manual ACKs**: Messages acknowledged only after successful processing.
- **Durable queues**: Messages survive broker restarts.
- **Idempotency**: Duplicate messages are safely handled and not re-processed.

## Architecture

See [architecture/architecture.md](../architecture/architecture.md) for a visual diagram.

**Event flow**:
```
1. Order Service calls Payment Service (gRPC).
2. Payment Service processes payment and commits to DB.
3. Payment Service publishes a JSON event to RabbitMQ queue "payment.completed".
4. Notification Service consumes the event.
5. Idempotency check: if message ID already processed, skip and ACK.
6. Otherwise: log notification (simulating email) and ACK.
```

## Implementation Details

### Idempotency Strategy

- **Location**: `notification-service/main.go` lines ~70–80.
- **Mechanism**: In-memory `map[string]struct{}` protected by mutex.
- **Fallback**: If message has no `MessageId`, use the entire message body as key.
- **Behavior**: On duplicate, ACK is sent but log is skipped. This prevents duplicate notifications.

### Manual ACKs & Reliability

- **Queue Declaration** (`notification-service/main.go` lines ~40–50):
  - `durable: true` → queue persists after broker restart.
  - `autoAck: false` → manual acknowledgments required.
- **Publisher** (`payment-service/internal/messaging/rabbitmq_publisher.go`):
  - Messages published with `DeliveryMode: Persistent`.
  - Ensures message survives broker failure.
- **Consumer** (`notification-service/main.go` lines ~85–110):
  - `Ack()` called only after log succeeds.
  - If consumer crashes before ACK, message remains in queue and is retried.

### Graceful Shutdown

- **Notification Service**: Uses `os/signal.NotifyContext()` to handle SIGTERM and SIGINT.
- **Payment Service**: Closes RabbitMQ publisher connection cleanly (can be extended).

## Running with Docker Compose

```bash
cd /path/to/AP2_Assignment1_name_surname_group
docker compose up --build
```

**Services**:
- `order-service` (HTTP: 8080, gRPC: 50051)
- `payment-service` (HTTP: 8081, gRPC: 50052)
- `notification-service` (logs only, no ports)
- `rabbitmq` (AMQP: 5672, Management: http://localhost:15672)
- `order-db` (Postgres: 5433)
- `payment-db` (Postgres: 5434)

**Test the flow**:
```bash
# Create an order (which triggers payment via gRPC)
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","item_name":"Laptop","amount":50000}'

# Watch Notification Service logs for the email output:
# [Notification] Sent email to user@example.com for Order #<id>. Amount: $<amount>
```

## Project Files Added/Modified

### New Files
- `notification-service/main.go` — Consumer implementation with idempotency.
- `notification-service/go.mod` — Go module definition.
- `notification-service/Dockerfile` — Multi-stage build.
- `payment-service/internal/messaging/rabbitmq_publisher.go` — Publisher implementation.

### Modified Files
- `docker-compose.yml` — Added `rabbitmq` and `notification-service`.
- `payment-service/internal/usecase/payment_usecase.go` — Added event publishing after DB commit.
- `payment-service/cmd/payment-service/main.go` — Initialize publisher.
- `payment-service/go.mod` — Added `github.com/rabbitmq/amqp091-go`.

## Notes

- Idempotency is in-memory; for production, use a persistent store (Postgres table).
- Email is a placeholder (`user@example.com`); in production, extract from Order/Payment.
- DLQ (Dead Letter Queue) bonus: Not yet implemented; can be added for retry policy.
- All gRPC connections use insecure credentials (suitable for local/Docker networks).

## Submission Checklist

- ✅ Event-driven architecture implemented (RabbitMQ).
- ✅ Notification Service consumer with manual ACKs.
- ✅ Idempotency check in place.
- ✅ Durable queues and persistent delivery.
- ✅ Graceful shutdown in Notification Service.
- ✅ Docker Compose orchestration.
- ✅ Updated architecture diagram.
- ✅ README explaining design.