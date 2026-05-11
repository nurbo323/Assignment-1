# AP2 Assignment 1 / 4 - Clean Architecture Microservices

This repository contains a complete Go microservices solution with three processes:

- **Order Service** — manages orders, caching, and order lifecycle
- **Payment Service** — handles payment authorization and payment records
- **Notification Service** — consumes payment events and sends notifications

The project follows Clean Architecture, separates persistence per service, avoids shared domain models, and uses Docker Compose for local orchestration.

## Highlights

- **Clean Architecture:** thin HTTP handlers, application use cases, repositories, and pure domain logic
- **Service boundaries:** each service owns its own database and responsibilities
- **Order caching:** Redis cache-aside for fast order reads
- **Event-driven notifications:** RabbitMQ queue `payment.completed` drives the notification worker
- **Idempotency:** order submission supports `Idempotency-Key`, and notification processing is guarded by Redis locks
- **Local development:** one command starts the full stack with Docker Compose

## Architecture

![Architecture Diagram](./architecture/architecture.png)

The source diagram is also available in [architecture/architecture.md](./architecture/architecture.md) and [architecture/architecture.dot](./architecture/architecture.dot).

## Repository layout

```text
.
├── docker-compose.yml
├── README.md
├── architecture/
│   ├── architecture.md
│   ├── architecture.dot
│   └── architecture.png
├── order-service/
├── payment-service/
├── notification-service/
├── generated-contracts/
├── proto-repository/
└── tools/
```

## Service responsibilities

### Order Service

Owns:

- create order
- read order
- cancel order
- order status transitions
- cache maintenance in Redis

Does not own:

- payment authorization rules
- payment persistence

### Payment Service

Owns:

- process payment
- persist payment result
- generate transaction metadata
- reject payments above the configured threshold

Does not own:

- order lifecycle decisions

### Notification Service

Owns:

- consume `payment.completed` events
- avoid duplicate notification delivery
- simulate or send real email notifications
- retry failed delivery with exponential backoff

## Business rules

- Monetary values use `int64`
- Order amount must be greater than zero
- `Paid` orders cannot be cancelled
- Payments above `100000` are declined
- Order Service uses an HTTP timeout of `2s` for the payment call

## How the system works

1. Client creates an order via `POST /orders`
2. Order Service stores the order and calls Payment Service
3. Payment Service creates a payment and publishes `payment.completed` to RabbitMQ
4. Notification Service consumes the message
5. Notification Service sends a notification and stores the result in Redis

## Run locally

### Start everything

```bash
docker compose up --build -d
```

### Check running containers

```bash
docker compose ps
```

### Watch logs

```bash
docker compose logs -f order-service payment-service notification-service rabbitmq redis
```

## API examples

### Create order

```bash
curl -i -X POST http://localhost:8080/orders \
    -H "Content-Type: application/json" \
    -H "Idempotency-Key: demo-key-1" \
    -d '{"customer_id":"cust-1","item_name":"Keyboard","amount":15000}'
```

### Get order

```bash
curl http://localhost:8080/orders/<order_id>
```

### Cancel order

```bash
curl -X PATCH http://localhost:8080/orders/<order_id>/cancel
```

### Order stats

```bash
curl http://localhost:8080/order/stats
```

### Create payment directly

```bash
curl -i -X POST http://localhost:8081/payments \
    -H "Content-Type: application/json" \
    -d '{"order_id":"<order_id>","amount":15000}'
```

### Get payment by order id

```bash
curl http://localhost:8081/payments/<order_id>
```

### Payment stats

```bash
curl -s http://localhost:8081/payments/stats
```

## gRPC verification

Payment Service also exposes gRPC internally on `:50051`.

If you want to test it from the host, expose the port in `docker-compose.yml` or run the request from a container in the same Docker network.

Example with `grpcurl`:

```bash
grpcurl -plaintext -d '{"order_id":"<order_id>","amount":15000}' localhost:50051 payment.v1.PaymentService/ProcessPayment
grpcurl -plaintext -d '{}' localhost:50051 payment.v1.PaymentService/GetPaymentStats
```

## Environment variables

### Order Service

- `APP_PORT` — default `8080`
- `DB_DSN` — default `postgres://postgres:postgres@localhost:5433/orders_db?sslmode=disable`
- `PAYMENT_BASE_URL` — default `http://localhost:8081`
- `REDIS_ADDR` — default `localhost:6379`
- `CACHE_TTL` — default `5m`

### Payment Service

- `APP_PORT` — default `8081`
- `DB_DSN` — default `postgres://postgres:postgres@localhost:5434/payments_db?sslmode=disable`
- `RABBITMQ_URL` — default `amqp://guest:guest@rabbitmq:5672/`

### Notification Service

- `RABBITMQ_URL` — default `amqp://guest:guest@rabbitmq:5672/`
- `REDIS_ADDR` — default `localhost:6379`
- `PROVIDER_MODE` — default `SIMULATED`
- `RETRY_MAX_ATTEMPTS` — default `3`
- `RETRY_BASE_DELAY` — default `2s`
- `RETRY_MAX_DELAY` — default `8s`
- `PROCESSING_LOCK_TTL` — default `30s`

## Verification checklist

- `docker compose ps` shows all services as `Up`
- `payment.completed` queue has `1` consumer
- `messages_ready = 0` and `messages_unacknowledged = 0`
- `notification-service` logs show worker startup and message handling
- Creating an order returns `201 Created`
- Payment lookup by `order_id` returns the created payment

## Submission

Recommended archive name:

```text
AP2_Assignment1_name_surname_group.zip
```

Upload the zip to Moodle after committing your changes to git.
