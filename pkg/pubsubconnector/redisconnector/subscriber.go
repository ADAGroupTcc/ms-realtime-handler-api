package redisconnector

import (
	"context"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
	redis "github.com/go-redis/redis/v8"
)

type redisSubscriber struct {
	client     *redis.Client
	log        *logger.Logger
	instrument interfaces.Instrument
}

func NewRedisSubscriber(config *redisConnectionConfig, log *logger.Logger, instrument interfaces.Instrument) pubsubconnector.Subscriber {
	return &redisSubscriber{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			PoolSize: config.PoolSize, // Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
		}),
		log:        log,
		instrument: instrument,
	}
}

func (redisSubscriber *redisSubscriber) SubscribeAsync(ctx context.Context, topic string, eventsChan chan []byte) {
	pubsub := redisSubscriber.client.Subscribe(ctx, topic)

	ch := pubsub.Channel()

	msg := <-ch
	eventsChan <- []byte(msg.Payload)
	redisSubscriber.log.Info("redis_subscriber: received message from topic", logger.WithEvent(logger.Event{
		"topic":   msg.Channel,
		"payload": msg.Payload,
	}))

	//for msg := range ch {
	//	eventsChan <- []byte(msg.Payload)
	//	redisSubscriber.log.Info("redis_subscriber: received message from topic", logger.WithEvent(logger.Event{
	//		"topic":   msg.Channel,
	//		"payload": msg.Payload,
	//	}))
	//}
}

// TODO: implementar observabilidade
