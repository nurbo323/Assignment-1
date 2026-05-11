package state

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

type Store struct {
	client    *redislib.Client
	lockTTL   time.Duration
	sentTTL   time.Duration
	failedTTL time.Duration
	prefix    string
}

func NewStore(client *redislib.Client, lockTTL, sentTTL, failedTTL time.Duration) *Store {
	return &Store{
		client:    client,
		lockTTL:   lockTTL,
		sentTTL:   sentTTL,
		failedTTL: failedTTL,
		prefix:    "notification:payment:",
	}
}

func (s *Store) GetStatus(ctx context.Context, paymentID string) (string, bool, error) {
	value, err := s.client.Get(ctx, s.statusKey(paymentID)).Result()
	if err == redislib.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (s *Store) TryLock(ctx context.Context, paymentID string) (bool, error) {
	return s.client.SetNX(ctx, s.lockKey(paymentID), "1", s.lockTTL).Result()
}

func (s *Store) MarkSent(ctx context.Context, paymentID string) error {
	return s.client.Set(ctx, s.statusKey(paymentID), "sent", s.sentTTL).Err()
}

func (s *Store) MarkFailed(ctx context.Context, paymentID string) error {
	return s.client.Set(ctx, s.statusKey(paymentID), "failed", s.failedTTL).Err()
}

func (s *Store) ReleaseLock(ctx context.Context, paymentID string) error {
	return s.client.Del(ctx, s.lockKey(paymentID)).Err()
}

func (s *Store) statusKey(paymentID string) string {
	return s.prefix + paymentID + ":status"
}

func (s *Store) lockKey(paymentID string) string {
	return s.prefix + paymentID + ":lock"
}
