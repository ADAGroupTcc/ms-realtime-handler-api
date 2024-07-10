package redisconnector

import (
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/cache"
	r "github.com/go-redis/redis"
)

type redisCache struct {
	client *r.Client
	log    *logger.Logger
}

// New Return new Redis cache
func NewCache(config *redisConnectionConfig, log *logger.Logger) cache.Cache {
	return &redisCache{
		client: r.NewClient(&r.Options{
			Addr:     config.Addr,
			Password: "",
			DB:       0,
		}),
		log: log,
	}
}

func (c *redisCache) Set(key string, value string) error {
	err := c.client.Set(key, value, 0).Err()
	c.log.Debugf("Redis - Set key %s", key)
	return err
}

func (c *redisCache) Get(key string) (string, error) {
	val, err := c.client.Get(key).Result()
	c.log.Debugf("Redis - Get key %s", key)
	if err == r.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return val, nil
	}
}

func (c *redisCache) Delete(key string) error {
	err := c.client.Del(key).Err()
	c.log.Debugf("Redis - Delete key %s", key)
	return err
}
