package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitURL := "amqp://guest:guest@rabbitmq:5672/"
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		rabbitURL = v
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	qName := "payment.completed"
	_, err = ch.QueueDeclare(
		qName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	// QoS: process one message at a time
	_ = ch.Qos(1, 0, false)

	msgs, err := ch.Consume(
		qName,
		"",
		false, // autoAck=false -> manual ACKs
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to register consumer: %v", err)
	}

	processed := make(map[string]struct{})
	var mu sync.Mutex

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("notification-service: waiting for messages")

	for {
		select {
		case <-ctx.Done():
			log.Println("notification-service: shutting down")
			return
		case d, ok := <-msgs:
			if !ok {
				log.Println("channel closed")
				return
			}

			id := d.MessageId
			if id == "" {
				// fallback: use body hash
				id = string(d.Body)
			}

			mu.Lock()
			_, seen := processed[id]
			mu.Unlock()

			if seen {
				// idempotency: ack and skip
				_ = d.Ack(false)
				continue
			}

			// simulate sending email by logging
			log.Printf("[Notification] Sent email for message id=%s payload=%s", id, string(d.Body))

			// mark processed and ACK
			mu.Lock()
			processed[id] = struct{}{}
			mu.Unlock()

			if err := d.Ack(false); err != nil {
				log.Printf("failed to ack message: %v", err)
			}
		}
	}
}
