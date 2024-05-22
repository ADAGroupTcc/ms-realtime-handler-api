package router

import (
	"context"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/clients"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/websocket"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/middlewares"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/logs"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
	"github.com/gin-gonic/gin"
)

type HandlersDependencies struct {
	PubSubBroker    *pubsubconnector.PubSubBroker
	SessionClienter clients.SessionClienter
	Instrument      interfaces.Instrument
}

func Handlers(ctx context.Context, dependencies *HandlersDependencies) *gin.Engine {
	jsonFormatter := &logs.JSONFormatter{}

	gi := gin.New()
	gi.Use(middlewares.SetRequestId())
	gi.Use(middlewares.EnhanceLogger(jsonFormatter))

	logger := logs.New(jsonFormatter)

	gi.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	websocketHandler := websocket.NewHandler(dependencies.PubSubBroker, dependencies.Instrument, logger)

	gi.Use(middlewares.Authenticate(dependencies.SessionClienter, logger))

	gi.GET("/ws", websocketHandler.WebsocketServer)

	return gi
}
