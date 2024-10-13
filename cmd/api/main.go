package main

import (
	"context"
	"os"
	"time"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/cache/rediscache"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/pubsubconnector/kafkaconnector"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/config"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/router"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/pubsubconnector"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/pubsubconnector/redisconnector"
	_ "go.uber.org/automaxprocs"
)

func main() {
	envs := config.LoadEnvVars()
	ctx := context.Background()

	redisPubSubConnectionconfig := redisconnector.NewConfig(
		envs.RedisHost,
		envs.RedisPoolSize,
	)

	redisCacheConnectionConfig := rediscache.NewConfig(envs.RedisHost)

	publisher, err := kafkaconnector.NewKafkaProducer(envs.KafkaBrokers)
	if err != nil {
		os.Exit(1)
	}
	subscriber := redisconnector.NewRedisSubscriber(redisPubSubConnectionconfig)
	broker := pubsubconnector.NewPubSubBroker(publisher, subscriber)
	cache := rediscache.NewCache(redisCacheConnectionConfig)

	subscribeEventChan := make(chan []byte, 100)

	wsConnectionsService := services.NewWebsocketConnectionsService(time.Duration(envs.WsReadDeadlineAwaitSeconds)*time.Second, cache)
	publishService := services.NewPublishEventService(broker, envs.KafkaPublisherTopic)
	subscribeService := services.NewSubscribeEventService(broker, envs.RedisSubscribeTopic, subscribeEventChan, wsConnectionsService)

	handlers := router.Handlers(ctx,
		&router.HandlersDependencies{
			PublishService:      publishService,
			SubscribeService:    subscribeService,
			WsConnectionService: wsConnectionsService,
		},
	)
	err = handlers.Run(":" + envs.APIPort)
	if err != nil {
		os.Exit(1)
	}
}
