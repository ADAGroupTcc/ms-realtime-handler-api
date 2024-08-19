package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/domain"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/services"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/util"
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
	instrument          interfaces.Instrument
	log                 *logger.Logger
}

func NewHandler(
	publishService services.PublishServicer,
	subscribeService services.SubscribeServicer,
	wsConnectionService services.WsConnectionServicer,
	instrument interfaces.Instrument,
	log *logger.Logger,
) *websocketHandler {
	return &websocketHandler{
		publishService:      publishService,
		subscribeService:    subscribeService,
		wsConnectionService: wsConnectionService,
		instrument:          instrument,
		log:                 log,
	}
}

func (h *websocketHandler) WebsocketServer(c *gin.Context) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error(util.FailedToUpgradeConnection, err)
		return
	}

	ctx := c.Request.Context()
	podName := os.Getenv("HOSTNAME")
	userId := c.Request.Header.Get("user_id")

	h.log.Debugf(util.UserIsConnected, userId, podName)

	h.wsConnectionService.SetConn(ctx, userId, conn)
	go h.wsConnectionService.RefreshConnection(ctx, userId)

	h.log.Debugf(util.NumberOfActiveConnections, h.wsConnectionService.ConnectionSize())

	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				h.deleteConn(ctx, userId, err, util.ConnectionClosedUnexpectedly)
				return
			}

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				h.deleteConn(ctx, userId, err, util.ConnectionClosed)
				return
			}

			h.deleteConn(ctx, userId, err, util.FailedToReadMessageFromWebsocket)
			return
		}

		eventReceived := domain.EventReceived{}
		err = json.Unmarshal(msg, &eventReceived)
		if err != nil {
			h.log.Error(util.FailedToUnmarshalMessage, err)
			activeConn := h.wsConnectionService.GetConn(userId)
			if activeConn != nil {
				sendEventError(activeConn.Conn, eventReceived.EventId, eventReceived.EventType, err, http.StatusBadRequest)
			}
			return
		}

		err = eventReceived.Validate()
		if err != nil {
			h.log.Error(util.FailedToValidateMessage, err)
			activeConn := h.wsConnectionService.GetConn(userId)
			if activeConn != nil {
				sendEventError(activeConn.Conn, eventReceived.EventId, eventReceived.EventType, err, http.StatusBadRequest)
			}
			return
		}

		eventToPublish := eventReceived.ToEventToPublish(userId)
		err = h.publishService.PublishEvent(ctx, eventToPublish, h.log)
		if err != nil {
			h.log.Error(util.FailedToPublishMessageToPubSubBroker, err)
			activeConn := h.wsConnectionService.GetConn(userId)
			sendEventError(activeConn.Conn, eventReceived.EventId, eventReceived.EventType, err, http.StatusInternalServerError)
			return
		}
		h.log.Infof(util.PublishMessageToPubSubBrokerSuccessfully, eventReceived.EventType)
	}
}

func (h *websocketHandler) deleteConn(ctx context.Context, userId string, err error, errDescr string) {
	h.wsConnectionService.DeleteConn(ctx, userId)
	h.log.Error(errDescr, err)
}

func sendEventError(conn *websocket.Conn, eventId string, eventName string, err error, code int) {
	conn.WriteJSON(map[string]interface{}{
		"event_id":   eventId,
		"event_type": eventName,
		"content": map[string]interface{}{
			"error": err.Error(),
			"code":  code,
		},
	})
}
