package websocket

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/helpers"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/services"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/cache"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/util"
	_ "github.com/PicPay/ms-chatpicpay-websocket-handler-api/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	mutex sync.Mutex

	activeConnections = helpers.NewActiveConnections()
)

type websocketHandler struct {
	publishService                            services.PublishServicer
	instrument                                interfaces.Instrument
	log                                       *logger.Logger
	cache                                     cache.Cache
	subscribeChan                             *chan []byte
	redisCacheConnectionExpirationTimeMinutes int
}

func NewHandler(
	publishService services.PublishServicer,
	instrument interfaces.Instrument,
	cache cache.Cache,
	log *logger.Logger,
	subscribeChan *chan []byte,
	redisCacheConnectionExpirationTimeMinutes int) *websocketHandler {
	return &websocketHandler{
		publishService: publishService,
		instrument:     instrument,
		log:            log,
		cache:          cache,
		subscribeChan:  subscribeChan,
		redisCacheConnectionExpirationTimeMinutes: redisCacheConnectionExpirationTimeMinutes,
	}
}

func handleSubscriptionResponse(subscriptionChan *chan []byte, podName string, log *logger.Logger) {
	mutex.Lock()
	defer mutex.Unlock()
	for subscribedEvent := range *subscriptionChan {
		subscribedEvent, err := parseEventToSendToReceiver(subscribedEvent)
		if err != nil {
			log.Error(util.UnableToParseEventResponse, err)
			continue
		}

		responseConn := activeConnections.GetConn(subscribedEvent.UserId)
		if responseConn == nil {
			log.Infof(util.ReceiverNotOnlineInPod, subscribedEvent.UserId, podName)
			continue
		}
		responseConn.WriteJSON(subscribedEvent)
		log.Infof("websocket_handler: event sent to receiver_id with data: %v", subscribedEvent.UserId, subscribedEvent.Data)
	}
}

func (h *websocketHandler) WebsocketServer(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error(util.FailedToUpgradeConnection, err)
		return
	}

	ctx := c.Request.Context()

	userId := c.Request.Header.Get("user_id")
	h.log.Debugf(util.UserIsConnected, userId)

	activeConnections.SetConn(userId, conn)
	podName := os.Getenv("HOSTNAME")
	h.cache.Set(userId, podName)

	h.log.Debugf(util.NumberOfActiveConnections, activeConnections.ConnectionSize())

	go handleSubscriptionResponse(h.subscribeChan, podName, h.log)

	for {
		_, msg, err := conn.ReadMessage()

		userConnTime := activeConnections.GetConnTime(userId)
		userConnTimeWithInterval := userConnTime.Add(time.Minute * time.Duration(h.redisCacheConnectionExpirationTimeMinutes))
		if time.Now().After(userConnTimeWithInterval) {
			h.cache.Delete(userId)
			h.cache.Set(userId, podName)
		}

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				h.deleteConn(userId, err, util.ConnectionClosedUnexpectedly)
				return
			}

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				h.deleteConn(userId, err, util.ConnectionClosed)
				return
			}

			h.deleteConn(userId, err, util.FailedToReadMessageFromWebsocket)
			return
		}

		eventReceived := EventReceived{}
		err = json.Unmarshal(msg, &eventReceived)
		if err != nil {
			sendEventError(
				conn, util.ErrorTypeErr, eventReceived.EventId, eventReceived.EventType, err, 400,
			)
			h.log.Error(util.FailedToUnmarshalMessage, err)
			return
		}

		err = eventReceived.Validate()
		if err != nil {
			sendEventError(
				conn, util.ErrorTypeErr, eventReceived.EventId, eventReceived.EventType, err, 400,
			)

			h.getActiveConn(userId, err, util.FailedToValidateMessage, 400)
			return
		}

		eventToPublish := eventReceived.ToEventToPublish(userId)
		err = h.publishService.PublishEvent(ctx, eventToPublish, h.log)

		if err != nil {
			h.getActiveConn(userId, err, util.FailedToPublishMessageToPubSubBroker, 500)
			return
		}
	}
}

func (h *websocketHandler) getActiveConn(userId string, err error, errDescr string, errorCode int) {
	h.log.Error(errDescr, err)
	activeConn := activeConnections.GetConn(userId)
	if activeConn != nil {
		sendError(activeConn, util.ErrorTypeErr, err, errorCode)
	}
}

func (h *websocketHandler) deleteConn(userId string, err error, errDescr string) {
	activeConnections.DeleteConn(userId)
	h.cache.Delete(userId)
	h.log.Error(errDescr, err)
}

func sendError(conn *websocket.Conn, errorType string, err error, code int) {
	conn.WriteJSON(map[string]interface{}{
		"type":  errorType,
		"error": err.Error(),
		"code":  code,
	})
}

func sendEventError(conn *websocket.Conn, errorType string, eventId string, eventName string, err error, code int) {
	conn.WriteJSON(map[string]interface{}{
		"event_id":   eventId,
		"event_name": eventName,
		"content": map[string]interface{}{
			"type":  errorType,
			"error": err.Error(),
			"code":  code,
		},
	})
}
