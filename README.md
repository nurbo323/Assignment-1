# AP2 Assignment 1 - Clean Architecture based Microservices (Order & Payment)

This solution implements the assignment as two separate Go microservices:
1. Order Service - owns order lifecycle and order state.
2. Payment Service - owns payment authorization and transaction records.

The implementation follows the assignment requirements from the PDF: Clean Architecture, REST-only communication through Gin, database per service, no shared entities/models, manual DI in `main.go`, and a custom HTTP client timeout of 2 seconds max for the Order -> Payment call.

## 1. How the solution maps to the assignment

### Clean Architecture inside each service
Each service is split into these layers:
- `internal/domain` - pure entities and domain errors/constants
- `internal/usecase` - business rules and application logic
- `internal/repository` - PostgreSQL persistence
- `internal/transport/http` - thin Gin handlers
- `cmd/<service-name>/main.go` - composition root / manual dependency injection

### Microservice decomposition
- Order Service has its own database and tables.
- Payment Service has its own database and tables.
- No service imports the other's entity/model package.
- Order Service communicates with Payment Service only via REST.

### Required business rules implemented
- Money uses `int64`, never `float64`.
- Order amount must be `> 0`.
- `Paid` orders cannot be cancelled.
- Payment amount over `100000` is declined.
- Order Service uses custom `http.Client{Timeout: 2 * time.Second}`.

### Failure-handling decision
If the Payment Service is unavailable:
- Order Service does not hang because it uses timeout.
- Order Service marks the order as Failed.
- Order Service returns 503 Service Unavailable.

I chose Failed instead of leaving the order Pending because:
- the request to create the order has already completed from the client's point of view,
- payment authorization did not succeed,
- the final state is easier to explain during defense,
- it avoids stuck pending orders.

### Bonus: simple idempotency
Bonus idempotency is included in the Order Service with the `Idempotency-Key` header.
- Same key + same request payload returns the already created order instead of creating duplicates.
- Payment Service also protects against duplicate payment records with a unique `order_id`.

## 2. Architecture diagram
![Architecture Diagram](./architecture/architecture.png)

Mermaid source is also included in `architecture/architecture.mmd`.

## 3. Project structure
```text
AP2_Assignment1_name_surname_group/
├── docker-compose.yml
├── README.md
├── architecture/
│   ├── architecture.mmd
│   └── architecture.png
├── order-service/
│   ├── cmd/order-service/main.go
│   ├── internal/
│   │   ├── app/
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── transport/http/
│   │   └── usecase/
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
└── payment-service/
    ├── cmd/payment-service/main.go
    ├── internal/
    │   ├── app/
    │   ├── domain/
    │   ├── repository/
    │   ├── transport/http/
    │   └── usecase/
    ├── migrations/
    ├── Dockerfile
    └── go.mod
```

## 4. Bounded contexts

### Order bounded context
Owns:
- order creation
- order status transitions
- cancellation policy
- idempotency for order submission

Does not own:
- payment authorization decision logic
- transaction storage

### Payment bounded context
Owns:
- payment authorization
- payment status
- transaction ID generation
- payment transaction persistence
- amount limit validation

Does not own:
- order lifecycle rules beyond `order_id` reference

## 5. API examples

### Create order
```bash
curl -i -X POST http://localhost:8080/orders   -H "Content-Type: application/json"   -H "Idempotency-Key: demo-key-1"   -d '{"customer_id":"cust-1","item_name":"Keyboard","amount":15000}'
```

### Get order
```bash
curl http://localhost:8080/orders/<order_id>
```

### Cancel order
```bash
curl -X PATCH http://localhost:8080/orders/<order_id>/cancel
```

### Create payment directly
```bash
curl -i -X POST http://localhost:8081/payments   -H "Content-Type: application/json"   -d '{"order_id":"order-1","amount":15000}'
```

### Get payment by order id
```bash
curl http://localhost:8081/payments/order-1
```

## 6. How to run locally

### Option A - with Docker Compose
From the root folder:
```bash
docker compose up --build
```

Then apply SQL migrations manually:

**Order DB**
```bash
psql "postgres://postgres:postgres@localhost:5433/orders_db?sslmode=disable" -f order-service/migrations/001_init.sql
```

**Payment DB**
```bash
psql "postgres://postgres:postgres@localhost:5434/payments_db?sslmode=disable" -f payment-service/migrations/001_init.sql
```

### Option B - run each service manually
Start two PostgreSQL databases, apply migrations, then:

**Payment Service**
```bash
cd payment-service
go run ./cmd/payment-service
```

**Order Service**
```bash
cd order-service
go run ./cmd/order-service
```

## 7. Environment variables

### Order Service
- `APP_PORT` default: `8080`
- `DB_DSN` default: `postgres://postgres:postgres@localhost:5433/orders_db?sslmode=disable`
- `PAYMENT_BASE_URL` default: `http://localhost:8081`

### Payment Service
- `APP_PORT` default: `8081`
- `DB_DSN` default: `postgres://postgres:postgres@localhost:5434/payments_db?sslmode=disable`

## 8. Defense notes
- handlers are thin
- use cases depend on interfaces
- repositories implement those interfaces
- domain has no dependency on Gin or SQL
- wiring is done in `main.go`
- different processes, different DBs, no shared models, REST only

### Failure scenario
If Payment Service is down:
1. Order is inserted as `Pending`
2. Order Service tries `POST /payments`
3. HTTP client times out within 2 seconds
4. Order status is updated to `Failed`
5. API returns `503`

### Cancellation rule
Only `Pending` orders can be cancelled.

## 9. Submission
Zip this folder as:
```text
AP2_Assignment1_name_surname_group.zip
```
Then upload it to Moodle.
