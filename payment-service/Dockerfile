FROM golang:1.24-alpine AS builder
WORKDIR /app/payment-service
COPY payment-service/go.mod ./
COPY generated-contracts ../generated-contracts
RUN go mod download
COPY payment-service . 
RUN go build -o /bin/payment-service ./cmd/payment-service

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/payment-service /app/payment-service
EXPOSE 8081
CMD ["/app/payment-service"]
