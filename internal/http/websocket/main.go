package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/domain"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/services"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/cache"
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
	publishService                            services.PublishServicer
	subscribeService                          services.SubscribeServicer
	wsConnectionService                       services.WsConnectionServicer
	instrument                                interfaces.Instrument
	cache                                     cache.Cache
	log                                       *logger.Logger
	redisCacheConnectionExpirationTimeMinutes int
}

func NewHandler(
	publishService services.PublishServicer,
	subscribeService services.SubscribeServicer,
	wsConnectionService services.WsConnectionServicer,
	instrument interfaces.Instrument,
	cache cache.Cache,
	log *logger.Logger,
	redisCacheConnectionExpirationTimeMinutes int) *websocketHandler {
	return &websocketHandler{
		publishService:      publishService,
		subscribeService:    subscribeService,
		wsConnectionService: wsConnectionService,
		instrument:          instrument,
		cache:               cache,
		log:                 log,
		redisCacheConnectionExpirationTimeMinutes: redisCacheConnectionExpirationTimeMinutes,
	}
}

func (h *websocketHandler) WebsocketServer(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error(util.FailedToUpgradeConnection, err)
		return
	}

	ctx := c.Request.Context()
	podName := os.Getenv("HOSTNAME")
	userId := c.Request.Header.Get("user_id")

	h.log.Debugf(util.UserIsConnected, userId)

	h.wsConnectionService.SetConn(userId, conn)
	h.cache.Set(ctx, userId, podName)

	h.log.Debugf(util.NumberOfActiveConnections, h.wsConnectionService.ConnectionSize())

	for {
		_, msg, err := conn.ReadMessage()
		h.refreshUserConnection(ctx, userId, podName, h.redisCacheConnectionExpirationTimeMinutes, h.cache, h.log)

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
	}
}

func (h *websocketHandler) deleteConn(ctx context.Context, userId string, err error, errDescr string) {
	h.wsConnectionService.DeleteConn(userId)
	h.cache.Delete(ctx, userId)
	h.log.Error(errDescr, err)
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

func (h *websocketHandler) refreshUserConnection(ctx context.Context, userId string, podName string, expirationTime int, cache cache.Cache, log *logger.Logger) {
	userConnTime := h.wsConnectionService.GetConnStartTime(userId)
	userConnTimeWithInterval := userConnTime.Add(time.Minute * time.Duration(expirationTime))
	if time.Now().After(userConnTimeWithInterval) {
		cache.Set(ctx, userId, podName)
	}
}
