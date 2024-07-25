package rediscache

import (
	"context"

	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/cache"
	"github.com/go-redis/redis/v8"
)

type redisConnectionConfig struct {
	Addr string
}

type redisCache struct {
	client *redis.Client
	log    *logger.Logger
}

func NewConfig(addr string) *redisConnectionConfig {
	return &redisConnectionConfig{
		Addr: addr,
	}
}

func NewCache(config *redisConnectionConfig, log *logger.Logger) cache.Cache {
	return &redisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: "",
			DB:       0,
		}),
		log: log,
	}
}

func (c *redisCache) Set(ctx context.Context, key string, value string) error {
	err := c.client.Set(ctx, key, value, 0).Err()
	c.log.Debugf("Redis - Set key %s", key)
	return err
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	c.log.Debugf("Redis - Get key %s", key)
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
	c.log.Debugf("Redis - Delete key %s", key)
	return err
}
