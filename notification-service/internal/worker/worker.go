package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"notification-service/internal/provider"
	"notification-service/internal/state"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Worker struct {
	conn      *amqp.Connection
	queueName string
	store     *state.Store
	sender    provider.Sender
	retries   int
	baseDelay time.Duration
	maxDelay  time.Duration
}

func New(conn *amqp.Connection, queueName string, store *state.Store, sender provider.Sender, retries int, baseDelay, maxDelay time.Duration) *Worker {
	return &Worker{
		conn:      conn,
		queueName: queueName,
		store:     store,
		sender:    sender,
		retries:   retries,
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
	}
}

type eventPayload struct {
	PaymentID     string `json:"payment_id"`
	OrderID       string `json:"order_id"`
	CustomerEmail string `json:"customer_email"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

func (w *Worker) Run(ctx context.Context) error {
	ch, err := w.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(w.queueName, true, false, false, false, nil); err != nil {
		return err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(w.queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Printf("notification-service worker listening on queue=%s", w.queueName)

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := w.handleDelivery(ctx, msg); err != nil {
				log.Printf("notification delivery failed: %v", err)
			}
		}
	}
}

func (w *Worker) handleDelivery(ctx context.Context, msg amqp.Delivery) error {
	var payload eventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		_ = msg.Ack(false)
		return fmt.Errorf("decode event: %w", err)
	}

	if payload.PaymentID == "" {
		payload.PaymentID = msg.MessageId
	}
	if payload.PaymentID == "" {
		payload.PaymentID = string(msg.Body)
	}
	if payload.CustomerEmail == "" {
		payload.CustomerEmail = "user@example.com"
	}

	status, found, err := w.store.GetStatus(ctx, payload.PaymentID)
	if err != nil {
		return err
	}
	if found && status == "sent" {
		return msg.Ack(false)
	}

	locked, err := w.store.TryLock(ctx, payload.PaymentID)
	if err != nil {
		return err
	}
	if !locked {
		return msg.Nack(false, true)
	}
	defer func() {
		_ = w.store.ReleaseLock(ctx, payload.PaymentID)
	}()

	notification := provider.Notification{
		PaymentID:     payload.PaymentID,
		OrderID:       payload.OrderID,
		CustomerEmail: payload.CustomerEmail,
		Amount:        payload.Amount,
		Status:        payload.Status,
	}

	for attempt := 1; attempt <= w.retries; attempt++ {
		sendErr := w.sender.Send(ctx, notification)
		if sendErr == nil {
			if err := w.store.MarkSent(ctx, payload.PaymentID); err != nil {
				return err
			}
			return msg.Ack(false)
		}

		if attempt == w.retries {
			if err := w.store.MarkFailed(ctx, payload.PaymentID); err != nil {
				return err
			}
			return msg.Ack(false)
		}

		delay := exponentialBackoff(w.baseDelay, w.maxDelay, attempt)
		log.Printf("notification retry payment=%s attempt=%d delay=%s err=%v", payload.PaymentID, attempt, delay, sendErr)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil
}

func exponentialBackoff(baseDelay, maxDelay time.Duration, attempt int) time.Duration {
	if baseDelay <= 0 {
		baseDelay = 2 * time.Second
	}
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
	if maxDelay > 0 && delay > maxDelay {
		return maxDelay
	}
	return delay
}
