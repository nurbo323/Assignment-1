package messaging

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewRabbitPublisher(url, queue string) (*RabbitPublisher, error) {
	conn, err := amqp.DialConfig(url, amqp.Config{Properties: amqp.Table{"connection_name": "payment-publisher"}})
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// declare durable queue
	_, err = ch.QueueDeclare(
		queue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitPublisher{conn: conn, channel: ch, queue: queue}, nil
}

func (r *RabbitPublisher) PublishPaymentCompleted(ctx context.Context, payload []byte) error {
	// ensure publish happens with persistent delivery
	err := r.channel.PublishWithContext(ctx,
		"",      // default exchange
		r.queue, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         payload,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		log.Printf("failed to publish message: %v", err)
	}
	return err
}

func (r *RabbitPublisher) Close() error {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
