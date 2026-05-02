FROM golang:1.20-alpine AS build
WORKDIR /src
COPY . .
RUN apk add --no-cache git gcc musl-dev
RUN go build -o /bin/notification-service ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/notification-service /bin/notification-service
CMD ["/bin/notification-service"]
