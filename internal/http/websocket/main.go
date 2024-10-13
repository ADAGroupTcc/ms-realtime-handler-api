package websocket

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/services"

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
	publishService      services.PublishServicer
	subscribeService    services.SubscribeServicer
	wsConnectionService services.WsConnectionServicer
}

func NewHandler(
	publishService services.PublishServicer,
	subscribeService services.SubscribeServicer,
	wsConnectionService services.WsConnectionServicer,
) *websocketHandler {
	return &websocketHandler{
		publishService:      publishService,
		subscribeService:    subscribeService,
		wsConnectionService: wsConnectionService,
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
		// Envia uma mensagem de fechamento para o cliente
		if err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Connection closed")); err != nil {
			// Se falhar ao enviar a mensagem de fechamento, registrar o erro
		}
		conn.Close() // Garante que a conexão WebSocket será fechada
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

		eventToPublish := eventReceived.ToEventToPublish(userId)
		err = h.publishService.PublishEvent(ctx, eventToPublish)
		if err != nil {
			activeConn := h.wsConnectionService.GetConn(userId)
			sendEventError(activeConn.Conn, eventReceived.EventId, eventReceived.EventType, err, http.StatusInternalServerError)
			return
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
