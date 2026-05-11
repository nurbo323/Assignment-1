package redis

import (
	"context"
	"encoding/json"
	"order-service/internal/domain"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

type OrderCache struct {
	client *redislib.Client
	ttl    time.Duration
	prefix string
}

func NewOrderCache(client *redislib.Client, ttl time.Duration) *OrderCache {
	return &OrderCache{client: client, ttl: ttl, prefix: "order:cache:"}
}

func (c *OrderCache) Get(ctx context.Context, id string) (domain.Order, bool, error) {
	value, err := c.client.Get(ctx, c.key(id)).Result()
	if err == redislib.Nil {
		return domain.Order{}, false, nil
	}
	if err != nil {
		return domain.Order{}, false, err
	}

	var order domain.Order
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return domain.Order{}, false, err
	}

	return order, true, nil
}

func (c *OrderCache) Set(ctx context.Context, order domain.Order) error {
	value, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.key(order.ID), value, c.ttl).Err()
}

func (c *OrderCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, c.key(id)).Err()
}

func (c *OrderCache) key(id string) string {
	return c.prefix + id
}
