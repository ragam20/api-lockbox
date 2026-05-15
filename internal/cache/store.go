package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	Set(ctx context.Context, key string, value string, exp time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetDel(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

type store struct {
	client *redis.Client
}

func NewStore(c *redis.Client) Store {
	return &store{
		client: c,
	}
}

func (c *store) Set(ctx context.Context, key string, value string, exp time.Duration) error {
	return c.client.Set(ctx, key, value, exp).Err()
}

func (c *store) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *store) GetDel(ctx context.Context, key string) (string, error) {
	return c.client.GetDel(ctx, key).Result()
}

func (c *store) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func IsMiss(err error) bool {
	return errors.Is(err, redis.Nil)
}
