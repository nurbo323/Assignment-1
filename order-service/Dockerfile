FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /bin/order-service ./cmd/order-service

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/order-service /app/order-service
EXPOSE 8080
CMD ["/app/order-service"]
