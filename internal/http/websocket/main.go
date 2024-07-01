package websocket

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/helpers"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/services"
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
	publishService   services.PublishServicer
	subscribeService services.SubscribeServicer
	instrument       interfaces.Instrument
	log              *logger.Logger
}

func NewHandler(publishService services.PublishServicer, subscribeService services.SubscribeServicer, instrument interfaces.Instrument, log *logger.Logger) *websocketHandler {
	return &websocketHandler{
		publishService:   publishService,
		subscribeService: subscribeService,
		instrument:       instrument,
		log:              log,
	}
}

func (h *websocketHandler) WebsocketServer(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("websocket_handler: failed to upgrade connection", err)
		return
	}

	ctx := c.Request.Context()

	userId := c.Request.Header.Get("user_id")
	h.log.Debugf("websocket_handler: user_id: %s connected", userId)

	activeConnections.SetConn(userId, conn)
	h.log.Debugf("websocket_handler: number of active connections: %d", activeConnections.ConnectionSize())

	subscribeEventChan := make(chan []byte)
	go h.subscribeService.SubscribeAsync(ctx, subscribeEventChan, h.log)
	go func() {
		for event := range subscribeEventChan {
			event, err := parseEventToSendToReceiver(event)
			if err != nil {
				h.log.Error("unable to parser eventToReceiver response", err)
				//	Dúvida: Deveria ignorar e continuar?
				continue
			}

			responseConn := activeConnections.GetConn(event.ReceiverId)
			if responseConn == nil {
				h.log.Infof("receiver_id %s is not online in this pod_name: %s", event.ReceiverId, os.Getenv("HOSTNAME"))
				continue
			}
			responseConn.WriteJSON(event)
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				h.log.Error("websocket_handler: connection closed unexpectedly", err)
				activeConnections.DeleteConn(userId)
				return
			}

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				h.log.Debug("websocket_handler: connection closed")
				activeConnections.DeleteConn(userId)

				return
			}

			activeConnections.DeleteConn(userId)
			h.log.Error("websocket_handler: failed to read message from webSocket client", err)
			return
		}

		eventReceived := EventReceived{}
		err = json.Unmarshal(msg, &eventReceived)
		if err != nil {
			h.log.Error("websocket_handler: failed to unmarshal message", err)
			/*
				Dúvida: Aqui devo postar no tópico de resposta do Redis como um erro?
				Ou devo retonar o erro pela própria conexão?
			*/
			return
		}

		err = eventReceived.Validate()
		if err != nil {
			/*
				Dúvida: Aqui devo postar no tópico de resposta do Redis como um erro?
				Ou devo retonar o erro pela própria conexão?
			*/
			h.log.Error("websocket_handler: failed to validate message", err)
			activeConn := activeConnections.GetConn(userId)
			if activeConn != nil {
				activeConn.WriteJSON(map[string]interface{}{
					"error": err.Error(),
				})
				sendError(activeConn, "error", err, 400)
			}
			return
		}

		eventToPublish := eventReceived.ToEventToPublish(userId)

		err = h.publishService.PublishEvent(ctx, eventToPublish, h.log)
		if err != nil {
			h.log.Error("websocket_handler: failed to publish message to pubsub broker", err)
			activeConn := activeConnections.GetConn(userId)
			if activeConn != nil {
				sendError(activeConn, "error", err, 500)
			}
			return
		}
	}
}

func sendError(conn *websocket.Conn, errorType string, err error, code int) {
	conn.WriteJSON(map[string]interface{}{
		"type":  errorType,
		"error": err.Error(),
		"code":  code,
	})
}
