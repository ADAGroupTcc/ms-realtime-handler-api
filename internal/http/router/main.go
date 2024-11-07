package router

import (
	"context"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/websocket"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services/events"

	"github.com/gin-gonic/gin"
)

type HandlersDependencies struct {
	WsConnectionService services.WsConnectionServicer
	MessageSent         events.Services
	SearchRequested     events.Services
	ChannelAccepted     events.Services
	ChannelRejected     events.Services
}

func Handlers(ctx context.Context, dependencies *HandlersDependencies) *gin.Engine {
	gi := gin.New()

	websocketHandler := websocket.NewHandler(
		dependencies.WsConnectionService,
		map[string]events.Services{
			"MESSAGE_SENT":          dependencies.MessageSent,
			domain.SEARCH_REQUESTED: dependencies.SearchRequested,
			domain.CHANNEL_ACCEPTED: dependencies.ChannelAccepted,
			domain.CHANNEL_REJECTED: dependencies.ChannelRejected,
		},
	)

	gi.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	gi.GET("/ws", websocketHandler.WebsocketServer)

	return gi
}
