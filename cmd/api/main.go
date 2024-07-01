package main

import (
	"context"
	"os"
	"time"

	api "github.com/PicPay/lib-go-api"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/config"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/clients"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/router"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/services"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/http"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/instrumentation"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector/kafkaconnector"
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

	redisConnectionconfig := redisconnector.NewConfig(
		envs.RedisHost,
		envs.RedisPoolSize,
	)

	publisher, _ := kafkaconnector.NewKafkaProducer(envs.KafkaBrokers, log, instrument)
	subscriber := redisconnector.NewRedisSubscriber(redisConnectionconfig, log, instrument)
	broker := pubsubconnector.NewPubSubBroker(publisher, subscriber)

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

	sessionClient := clients.NewSessionClient(httpClient, instrument)

	if err != nil {
		log.Fatal("Failed to create http client", err)
		os.Exit(1)
	}

	publishService := services.NewPublishEventService(broker, envs.KafkaPublisherTopic)
	subscribeService := services.NewSubscribeEventService(broker, envs.RedisSubscribeTopic)

	handlers := router.Handlers(ctx,
		&router.HandlersDependencies{
			PublishService:   publishService,
			SubscribeService: subscribeService,
			SessionClienter:  sessionClient,
			Instrument:       instrument,
		},
	)

	err = api.Start(log, envs.APIPort, handlers)
	if err != nil {
		log.Fatal("Failed to start server", err)
	}
}
