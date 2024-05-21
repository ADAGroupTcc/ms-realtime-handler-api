package redisconnector

import (
	"context"
	"fmt"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"

	redis "github.com/go-redis/redis/v8"
)

type RedisConnectionConfig struct {
	Addr     string
	PoolSize int
}

type redisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(config *RedisConnectionConfig) *redisPublisher {
	return &redisPublisher{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			PoolSize: config.PoolSize, // Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
		}),
	}
}

func (redisPublisher *redisPublisher) Publisher(ctx context.Context, channel string, message string) error {
	resp := redisPublisher.client.Publish(ctx, channel, message)
	fmt.Println("Message published to channel:", resp)

	return nil
}

type redisSubscriber struct {
	client *redis.Client
}

func NewRedisSubscriber(config *RedisConnectionConfig) *redisSubscriber {
	return &redisSubscriber{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			PoolSize: config.PoolSize, // Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
		}),
	}
}

func (redisSubscriber *redisSubscriber) Subscriber(ctx context.Context, channel string) (*pubsubconnector.Event, error) {
	pubsub := redisSubscriber.client.Subscribe(ctx, channel)
	message, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println("======================| SUBSCRIBER | ======================")
	fmt.Println("Messsage: ", message.Payload)
	fmt.Println("Channel: ", message.Channel)
	fmt.Println("Pattern: ", message.Pattern)

	return &pubsubconnector.Event{
		Type:  message.Channel,
		Event: message.Payload,
	}, nil

	// ch := pubsub.Channel()
	// for msg := range ch {
	// 	fmt.Println("Received message from channel:", msg.Channel, "Message:", msg.Payload)
	// }
}
