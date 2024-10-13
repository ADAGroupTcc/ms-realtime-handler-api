package redisconnector

import (
	"context"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/pubsubconnector"

	redis "github.com/go-redis/redis/v8"
)

type redisSubscriber struct {
	client *redis.Client
}

func NewRedisSubscriber(config *redisConnectionConfig) pubsubconnector.Subscriber {
	return &redisSubscriber{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			PoolSize: config.PoolSize,
		}),
	}
}

func (redisSubscriber *redisSubscriber) SubscribeAsync(ctx context.Context, topic string, eventsChan chan []byte) {
	pubsub := redisSubscriber.client.Subscribe(ctx, topic)

	ch := pubsub.Channel()

	for msg := range ch {
		eventsChan <- []byte(msg.Payload)
	}
}
