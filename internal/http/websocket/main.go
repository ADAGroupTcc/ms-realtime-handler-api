package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services/events"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type websocketHandler struct {
	wsConnectionService services.WsConnectionServicer
	services            map[string]events.Services
}

func NewHandler(
	wsConnectionService services.WsConnectionServicer,
	services map[string]events.Services,
) *websocketHandler {
	return &websocketHandler{
		wsConnectionService: wsConnectionService,
		services:            services,
	}
}
func (h *websocketHandler) WebsocketServer(c *gin.Context) {
	userId := c.Request.Header.Get("user_id")
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer func() {
		if err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Connection closed")); err != nil {
			fmt.Println(err.Error())
		}
		conn.Close()
	}()

	ctx := c.Request.Context()
	h.wsConnectionService.SetConn(ctx, userId, conn)
	go h.wsConnectionService.RefreshConnection(ctx, userId)

	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.deleteConn(ctx, userId)
				return
			}

			h.deleteConn(ctx, userId)
			return
		}

		eventReceived := domain.EventReceived{}
		err = json.Unmarshal(msg, &eventReceived)
		if err != nil {
			activeConn := h.wsConnectionService.GetConn(userId)
			if activeConn != nil {
				sendEventError(activeConn.Conn, eventReceived.EventId, eventReceived.EventType, err, http.StatusBadRequest)
			}
			return
		}

		err = eventReceived.Validate()
		if err != nil {
			activeConn := h.wsConnectionService.GetConn(userId)
			if activeConn != nil {
				sendEventError(activeConn.Conn, eventReceived.EventId, eventReceived.EventType, err, http.StatusBadRequest)
			}
			return
		}

		service, ok := h.services[eventReceived.EventType]
		if !ok {
			activeConn := h.wsConnectionService.GetConn(userId)
			if activeConn != nil {
				sendEventError(activeConn.Conn, eventReceived.EventId, eventReceived.EventType, fmt.Errorf("event type not found"), http.StatusNotFound)
			}
			return
		}

		eventToPublish := eventReceived.ToEventToPublish(userId)
		eventBytes, err := json.Marshal(eventToPublish)
		if err != nil {
			activeConn := h.wsConnectionService.GetConn(userId)
			if activeConn != nil {
				sendEventError(activeConn.Conn, eventReceived.EventId, eventReceived.EventType, err, http.StatusInternalServerError)
			}
			return
		}

		eventsToPublish := service.Handle(ctx, eventBytes)

		for _, event := range eventsToPublish {
			activeConn := h.wsConnectionService.GetConn(event.UserId)
			if activeConn != nil {
				activeConn.Conn.WriteJSON(event)
			}
		}
	}
}

func (h *websocketHandler) deleteConn(ctx context.Context, userId string) {
	h.wsConnectionService.DeleteConn(ctx, userId)
}

func sendEventError(conn *websocket.Conn, eventId string, eventName string, err error, code int) {
	conn.WriteJSON(map[string]interface{}{
		"event_id":   eventId,
		"event_name": eventName,
		"content": map[string]interface{}{
			"error": err.Error(),
			"code":  code,
		},
	})
}
