FROM golang:1.24-alpine AS build
WORKDIR /app/notification-service
COPY notification-service/go.mod ./
COPY notification-service/go.sum ./
RUN go mod download
COPY notification-service . 
RUN apk add --no-cache git gcc musl-dev
RUN go build -o /bin/notification-service ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/notification-service /bin/notification-service
CMD ["/bin/notification-service"]
