FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /bin/payment-service ./cmd/payment-service

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/payment-service /app/payment-service
EXPOSE 8081
CMD ["/app/payment-service"]
