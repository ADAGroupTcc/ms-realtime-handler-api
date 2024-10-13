package rediscache

import (
	"context"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/cache"
	"github.com/go-redis/redis/v8"
)

type redisConnectionConfig struct {
	Addr string
}

type redisCache struct {
	client *redis.Client
}

func NewConfig(addr string) *redisConnectionConfig {
	return &redisConnectionConfig{
		Addr: addr,
	}
}

func NewCache(config *redisConnectionConfig) cache.Cache {
	return &redisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: "",
			DB:       0,
		}),
	}
}

func (c *redisCache) Set(ctx context.Context, key string, value string) error {
	err := c.client.Set(ctx, key, value, 0).Err()
	return err
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return val, nil
	}
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	return err
}
