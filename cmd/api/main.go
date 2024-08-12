package main

import (
	"context"
	"os"
	"time"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/cache/rediscache"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector/kafkaconnector"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/util"

	api "github.com/PicPay/lib-go-api"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/config"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/clients/sessionClient"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/router"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/services"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/http"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/instrumentation"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector/redisconnector"
	_ "go.uber.org/automaxprocs"
)

func main() {
	log := logger.New(
		logger.WithFatalHook(logger.WriteThenFatal),
	)

	envs := config.LoadEnvVars(log)
	ctx := context.Background()

	instrument := instrumentation.New(instrumentation.Config{
		Context:        ctx,
		AppName:        envs.AppName,
		AppEnv:         envs.AppEnv,
		Logger:         log,
		TraceEndpoint:  envs.InstrumentationTracesEndpoint,
		MetricEndpoint: envs.InstrumentationMetricsEndpoint,
	})

	redisPubSubConnectionconfig := redisconnector.NewConfig(
		envs.RedisHost,
		envs.RedisPoolSize,
	)

	redisCacheConnectionConfig := rediscache.NewConfig(envs.RedisHost)

	publisher, _ := kafkaconnector.NewKafkaProducer(envs.KafkaBrokers, log, instrument)
	subscriber := redisconnector.NewRedisSubscriber(redisPubSubConnectionconfig, log, instrument)
	broker := pubsubconnector.NewPubSubBroker(publisher, subscriber)
	cache := rediscache.NewCache(redisCacheConnectionConfig, log)

	httpClient, err := http.New(http.Config{
		BaseURL:         envs.SessionTokenAPIBaseURL,
		Timeout:         time.Duration(time.Millisecond * time.Duration(envs.SessionTokenAPITimeoutMs)),
		MaxIdleConns:    envs.SessionTokenMaxIdleConns,
		MaxConnsPerHost: envs.SessionTokenMaxConnsPerHost,
		RetryConfig: http.RetryConfig{
			Retries:         envs.SessionTokenAPIRetryCount,
			RetryAfter:      time.Duration(time.Duration(envs.SessionTokenAPIRetryIntervalMs) * time.Millisecond),
			RetryWhenStatus: envs.SessionTokenAPIRetryStatusCodes,
		},
		Logger:     log,
		Instrument: instrument,
	})

	sessionClient := sessionClient.NewSessionClient(httpClient, instrument)

	if err != nil {
		log.Fatal(util.FailedToCreateHttpClient, err)
		os.Exit(1)
	}

	subscribeEventChan := make(chan []byte, 100)

	wsConnectionsService := services.NewWebsocketConnectionsService(time.Duration(envs.WsReadDeadlineAwaitSeconds)*time.Second, cache, log)
	publishService := services.NewPublishEventService(broker, envs.KafkaPublisherTopic)
	subscribeService := services.NewSubscribeEventService(broker, envs.RedisSubscribeTopic, subscribeEventChan, wsConnectionsService)

	handlers := router.Handlers(ctx,
		&router.HandlersDependencies{
			PublishService:      publishService,
			SubscribeService:    subscribeService,
			WsConnectionService: wsConnectionsService,
			SessionClienter:     sessionClient,
			Instrument:          instrument,
		},
	)

	err = api.Start(log, envs.APIPort, handlers)
	if err != nil {
		log.Fatal(util.FailedToStartServer, err)
	}
}
