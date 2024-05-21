package main

import (
	"context"
	"fmt"
	"net/http"

	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/config"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/websocket"
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

	redisConnectionconfig := &redisconnector.RedisConnectionConfig{
		Addr:     envs.RedisHost,
		PoolSize: envs.RedisPoolSize,
	}
	publisher := redisconnector.NewRedisPublisher(redisConnectionconfig)
	subscriber := redisconnector.NewRedisSubscriber(redisConnectionconfig)
	broker := pubsubconnector.NewPubSubBroker(publisher, subscriber)

	handler := websocket.NewHandler(broker, instrument)

	http.HandleFunc("/ws", handler.WebsocketServer())

	log.Info(fmt.Sprintf("Server started on port %s", envs.APIPort))
	err := http.ListenAndServe(":"+envs.APIPort, nil)
	if err != nil {
		log.Fatal("Failed to start server", err)
	}
}
