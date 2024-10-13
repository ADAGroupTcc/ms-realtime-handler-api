package router

import (
	"context"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/websocket"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services"

	"github.com/gin-gonic/gin"
)

type HandlersDependencies struct {
	PublishService      services.PublishServicer
	SubscribeService    services.SubscribeServicer
	WsConnectionService services.WsConnectionServicer
}

func Handlers(ctx context.Context, dependencies *HandlersDependencies) *gin.Engine {
	gi := gin.New()

	websocketHandler := websocket.NewHandler(
		dependencies.PublishService,
		dependencies.SubscribeService,
		dependencies.WsConnectionService,
	)

	gi.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	go dependencies.SubscribeService.SubscribeAsync(ctx)
	go dependencies.SubscribeService.HandleSubscriptionResponse()
	gi.GET("/ws", websocketHandler.WebsocketServer)

	return gi
}
