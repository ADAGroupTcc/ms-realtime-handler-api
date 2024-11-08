package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/cache/rediscache"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/http"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/config"
	messagesClient "github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/clients/messages"
	sorterApi "github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/clients/sorter"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/router"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services/events"
	_ "go.uber.org/automaxprocs"
)

func main() {
	envs := config.LoadEnvVars()
	ctx := context.Background()

	redisCacheConnectionConfig := rediscache.NewConfig(envs.RedisHost)

	cache := rediscache.NewCache(redisCacheConnectionConfig)

	wsConnectionsService := services.NewWebsocketConnectionsService(time.Duration(envs.WsReadDeadlineAwaitSeconds)*time.Second, cache)

	messagesHttpClient, err := http.New(http.Config{
		BaseURL:           envs.MessagesApiUrl,
		Timeout:           time.Second * 10,
		AllowEmptyBaseUrl: false,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sorterHttpClient, err := http.New(http.Config{
		BaseURL:           envs.SorterApiUrl,
		Timeout:           time.Second * 10,
		AllowEmptyBaseUrl: false,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	messagesApi := messagesClient.New(messagesHttpClient)
	sorterApi := sorterApi.New(sorterHttpClient)

	handlers := router.Handlers(ctx,
		&router.HandlersDependencies{
			WsConnectionService: wsConnectionsService,
			MessageSent:         events.NewMessageSent(messagesApi),
			SearchRequested:     events.NewSearchRequested(sorterApi),
			ChannelAccepted:     events.NewChannelEvents(domain.CHANNEL_ACCEPTED),
			ChannelRejected:     events.NewChannelEvents(domain.CHANNEL_REJECTED),
		},
	)
	err = handlers.Run(":" + envs.APIPort)
	if err != nil {
		os.Exit(1)
	}
}
