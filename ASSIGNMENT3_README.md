# Assignment 3 — Event-Driven Architecture (EDA) with Message Queues

## Overview

This assignment implements an **asynchronous notification system** using **RabbitMQ** as a message broker. The Payment Service publishes events after successful payment processing, and the Notification Service consumes them with full reliability guarantees.

### Key Features
- ✅ **At-least-once delivery**: Messages are never lost if a consumer crashes.
- ✅ **Manual ACKs**: Messages acknowledged only after successful processing.
- ✅ **Durable queues**: Messages survive broker restarts.
- ✅ **Idempotency**: Duplicate messages are safely handled and not re-processed.
- ✅ **Graceful shutdown**: Services shut down cleanly on SIGTERM.

## Event Flow

```
1. Order Service calls Payment Service (gRPC).
2. Payment Service processes payment and commits to DB.
3. Payment Service publishes JSON event to RabbitMQ queue "payment.completed".
4. Notification Service consumes event from queue.
5. Idempotency check: if message ID already processed → skip log, send ACK.
6. Otherwise: log notification (simulating email) and send manual ACK.
```

## Implementation Details

### Idempotency Strategy

**Location**: `notification-service/main.go` (lines ~70–80)

**Mechanism**:
- In-memory `map[string]struct{}` protected by `sync.Mutex`
- **Fallback**: If message has no `MessageId`, use entire message body as key
- **Behavior**: On duplicate message, ACK is sent but notification is NOT logged
- **Limitation**: In-memory store is lost on restart; for production use Postgres table

### Manual ACKs & Reliability

**Queue Declaration** (`notification-service/main.go` lines ~40–50):
```go
_, err = ch.QueueDeclare(
    qName,
    true,  // durable: queue persists after broker restart
    false, // not auto-delete
    false, // not exclusive
    false, // no-wait
    nil,
)
```

**Consumer Configuration** (`notification-service/main.go`):
```go
msgs, err := ch.Consume(
    qName,
    "",
    false, // autoAck=false ⇒ manual acknowledgments
    false,
    false,
    false,
    nil,
)
```

**Publisher** (`payment-service/internal/messaging/rabbitmq_publisher.go`):
```go
err := r.channel.PublishWithContext(ctx,
    "",       // default exchange
    r.queue,  // routing key
    false,
    false,
    amqp.Publishing{
        ContentType:  "application/json",
        Body:         payload,
        DeliveryMode: amqp.Persistent, // ⇒ survives broker restart
        Timestamp:    time.Now(),
    },
)
```

### Graceful Shutdown

**Notification Service** (`notification-service/main.go`):
- Uses `os/signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`
- On signal, loop breaks and exits cleanly
- Connections to RabbitMQ are closed automatically on process exit

**Payment Service**: 
- Publisher exposes `Close()` method (can be called on shutdown)
- Currently initializes but doesn't call on exit (can be extended)

## Running with Docker Compose

### Quick Start

```bash
cd c:\Users\nurbo\Downloads\AP2_Assignment1_Nurbol_Amangeldinov_SE-2406\AP2_Assignment1_name_surname_group
docker compose up --build
```

### Services

| Service | Purpose | Port |
|---------|---------|------|
| `order-service` | Order processing | 8080 (HTTP), 50051 (gRPC) |
| `payment-service` | Payment processing, RabbitMQ publisher | 8081 (HTTP), 50052 (gRPC) |
| `notification-service` | Async notifications, RabbitMQ consumer | (no exposed ports) |
| `rabbitmq` | Message broker | 5672 (AMQP), 15672 (Management UI) |
| `order-db` | PostgreSQL for orders | 5433 |
| `payment-db` | PostgreSQL for payments | 5434 |

### Test the Event Flow

1. **Create an order** (triggers payment processing):
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id":"cust-1",
    "item_name":"Laptop",
    "amount":50000
  }'
```

2. **Observe notification logs**:
```bash
docker logs notification-service
# Expected output:
# [Notification] Sent email to user@example.com for Order #<id>. Amount: $<amount>
```

3. **Test idempotency** (retry with same order_id): Same event won't be logged twice.

4. **RabbitMQ Management UI**: `http://localhost:15672` (guest/guest)

## Project Structure

### New Files
- `notification-service/main.go` — Consumer with idempotency
- `notification-service/go.mod` — Go module
- `notification-service/Dockerfile` — Multi-stage build
- `payment-service/internal/messaging/rabbitmq_publisher.go` — Publisher
- `ASSIGNMENT3_README.md` — This file

### Modified Files
- `docker-compose.yml` — Added RabbitMQ and notification-service
- `payment-service/internal/usecase/payment_usecase.go` — Event publishing after DB commit
- `payment-service/cmd/payment-service/main.go` — Publisher initialization
- `payment-service/go.mod` — Added `github.com/rabbitmq/amqp091-go`
- `architecture/architecture.md` — Updated diagram with event flow

## Git Workflow for Submission

```bash
# Navigate to repo
cd c:\Users\nurbo\Downloads\AP2_Assignment1_Nurbol_Amangeldinov_SE-2406\AP2_Assignment1_name_surname_group

# Create feature branch
git checkout -b feature/assignment-3

# Stage and commit all changes
git add .
git commit -m "feat(assignment-3): add Event-Driven Architecture with RabbitMQ

- Implement Notification Service consumer (manual ACKs, idempotency)
- Add RabbitMQ publisher to Payment Service
- Update docker-compose with broker and notification service
- Ensure durable queues and at-least-once delivery
- Add graceful shutdown handling
- Update architecture diagram and README"

# Push to GitHub
git push origin feature/assignment-3
```

## Creating Submission ZIP

```bash
# From project parent directory
cd C:\Users\nurbo\Downloads\AP2_Assignment1_Nurbol_Amangeldinov_SE-2406

# Create ZIP archive with the entire project directory
# Windows PowerShell:
Compress-Archive -Path AP2_Assignment1_name_surname_group -DestinationPath AP2_Assignment3_Nurbol_Amangeldinov_SE-2406.zip

# or use 7-Zip / WinRAR context menu
```

**File to upload to Moodle**: `AP2_Assignment3_Nurbol_Amangeldinov_SE-2406.zip`

## Grading Checklist

| Criterion | Weight | Status |
|-----------|--------|--------|
| Messaging Logic (Producer/Consumer) | 30% | ✅ Implemented |
| Reliability & ACKs | 20% | ✅ Manual ACKs, durable queues |
| Idempotency | 20% | ✅ In-memory map check |
| Docker & Lifecycle | 20% | ✅ Graceful shutdown, full docker-compose |
| Documentation | 10% | ✅ README + updated diagram |

## Notes & Limitations

1. **Idempotency Storage**: Currently in-memory. For production, persist to `processed_messages` table in Postgres.
2. **Email Field**: Notification uses placeholder `user@example.com`. In production, extract from Order service.
3. **DLQ (Bonus)**: Dead Letter Queue not yet implemented. Can be added for retry policy.
4. **gRPC Credentials**: All connections use insecure credentials (suitable for local/Docker).
5. **API Version Warning**: If `docker compose config` warns about obsolete `version`, it's harmless (already removed).

## Next Steps on Defense

- Download ZIP from Moodle
- Run `docker compose up --build` (should work on any machine with Docker)
- Execute test curl commands and show logs
- Explain idempotency mechanism and ACK flow
- Answer questions about reliability guarantees

---

**Assignment 3 Complete** ✅
