package router

import (
	"context"
	"os"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/cache"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/clients/sessionClient"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/websocket"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/middlewares"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/services"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/logs"
	"github.com/gin-gonic/gin"
)

type HandlersDependencies struct {
	PublishService                            services.PublishServicer
	SubscribeService                          services.SubscribeServicer
	WsConnectionService                       services.WsConnectionServicer
	SessionClienter                           sessionClient.SessionClienter
	Instrument                                interfaces.Instrument
	Cache                                     cache.Cache
	RedisCacheConnectionExpirationTimeMinutes int
}

func Handlers(ctx context.Context, dependencies *HandlersDependencies) *gin.Engine {
	jsonFormatter := &logs.JSONFormatter{}

	gi := gin.New()
	gi.Use(middlewares.SetRequestId())
	gi.Use(middlewares.EnhanceLogger(jsonFormatter))

	logger := logs.New(jsonFormatter)

	websocketHandler := websocket.NewHandler(
		dependencies.PublishService,
		dependencies.SubscribeService,
		dependencies.WsConnectionService,
		dependencies.Instrument,
		dependencies.Cache,
		logger,
		dependencies.RedisCacheConnectionExpirationTimeMinutes)

	gi.Use(middlewares.Authenticate(dependencies.SessionClienter, logger))

	go dependencies.SubscribeService.SubscribeAsync(ctx, logger)
	go dependencies.SubscribeService.HandleSubscriptionResponse(os.Getenv("HOSTNAME"), logger)

	gi.GET("/ws", websocketHandler.WebsocketServer)

	gi.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return gi
}
