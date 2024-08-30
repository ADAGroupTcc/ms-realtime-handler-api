package router

import (
	"context"
	"os"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/clients/sessionClient"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/websocket"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/middlewares"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/services"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/logs"
	"github.com/gin-gonic/gin"
)

type HandlersDependencies struct {
	PublishService      services.PublishServicer
	SubscribeService    services.SubscribeServicer
	WsConnectionService services.WsConnectionServicer
	SessionClienter     sessionClient.SessionClienter
	Instrument          interfaces.Instrument
}

func Handlers(ctx context.Context, dependencies *HandlersDependencies) *gin.Engine {
	jsonFormatter := &logs.JSONFormatter{}

	gi := gin.New()
	gi.Use(middlewares.SetRequestId())
	gi.Use(middlewares.EnhanceLogger(jsonFormatter))

	logger := logs.New(jsonFormatter)
	gi.Use(middlewares.GetUserIdFromHeader(logger))
	gi.Use(middlewares.RecoverMiddleware(ctx, dependencies.Instrument, logger))

	websocketHandler := websocket.NewHandler(
		dependencies.PublishService,
		dependencies.SubscribeService,
		dependencies.WsConnectionService,
		dependencies.Instrument,
		logger,
	)

	gi.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	gi.GET("/panic", func(c *gin.Context) {
		panic("panic")
	})

	go dependencies.SubscribeService.SubscribeAsync(ctx, logger)
	go dependencies.SubscribeService.HandleSubscriptionResponse(os.Getenv("HOSTNAME"), logger)
	gi.GET("/ws", websocketHandler.WebsocketServer)

	return gi
}
