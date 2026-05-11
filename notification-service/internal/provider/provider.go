package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"strings"
	"time"

	"notification-service/internal/app"
)

type Notification struct {
	PaymentID     string
	OrderID       string
	CustomerEmail string
	Amount        int64
	Status        string
}

type Sender interface {
	Send(ctx context.Context, notification Notification) error
}

func NewSender(cfg app.Config) (Sender, error) {
	switch strings.ToUpper(cfg.ProviderMode) {
	case "REAL":
		return NewSMTPSender(cfg)
	default:
		return NewSimulatedSender(cfg), nil
	}
}

type SimulatedSender struct {
	minLatency  time.Duration
	failureRate float64
	rnd         *rand.Rand
}

func NewSimulatedSender(cfg app.Config) *SimulatedSender {
	return &SimulatedSender{
		minLatency:  cfg.SimulatedLatency,
		failureRate: cfg.SimulatedFailureRate,
		rnd:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *SimulatedSender) Send(ctx context.Context, notification Notification) error {
	delay := s.minLatency
	if delay <= 0 {
		delay = 300 * time.Millisecond
	}
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	if s.failureRate > 0 && s.rnd.Float64() < s.failureRate {
		return errors.New("simulated provider failure")
	}

	log.Printf("[Notification] sent email to %s for payment=%s order=%s amount=%d status=%s", notification.CustomerEmail, notification.PaymentID, notification.OrderID, notification.Amount, notification.Status)
	return nil
}

type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
	addr     string
	auth     smtp.Auth
}

func NewSMTPSender(cfg app.Config) (*SMTPSender, error) {
	if cfg.SMTPHost == "" || cfg.SMTPFrom == "" {
		return nil, errors.New("smtp configuration is incomplete")
	}

	addr := cfg.SMTPHost
	if cfg.SMTPPort != "" {
		addr = fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)
	}

	var auth smtp.Auth
	if cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	}

	return &SMTPSender{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		username: cfg.SMTPUsername,
		password: cfg.SMTPPassword,
		from:     cfg.SMTPFrom,
		addr:     addr,
		auth:     auth,
	}, nil
}

func (s *SMTPSender) Send(ctx context.Context, notification Notification) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	message := []byte(fmt.Sprintf("To: %s\r\nSubject: Payment %s confirmed\r\n\r\nYour payment %s for order %s in amount %d has been processed with status %s.\r\n", notification.CustomerEmail, notification.PaymentID, notification.PaymentID, notification.OrderID, notification.Amount, notification.Status))
	return smtp.SendMail(s.addr, s.auth, s.from, []string{notification.CustomerEmail}, message)
}
