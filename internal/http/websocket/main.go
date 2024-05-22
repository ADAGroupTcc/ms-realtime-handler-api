package websocket

import (
	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/helpers"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	activeConnections = helpers.NewActiveConnections()
)

type websocketHandler struct {
	broker     *pubsubconnector.PubSubBroker
	instrument interfaces.Instrument
	log        *logger.Logger
}

func NewHandler(broker *pubsubconnector.PubSubBroker, instrument interfaces.Instrument, log *logger.Logger) *websocketHandler {
	return &websocketHandler{
		broker:     broker,
		instrument: instrument,
		log:        log,
	}
}

func (h *websocketHandler) WebsocketServer(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("websocket_handler: failed to upgrade connection", err)
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()

	userId := c.Request.Header.Get("user_id")
	h.log.Debugf("websocket_handler: user_id: %s connected", userId)

	activeConnections.SetConn(userId, conn)
	h.log.Debugf("websocket_handler: number of active connections: %d", activeConnections.ConnectionSize())

	go func() {
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

			if err := h.broker.Publisher.Publisher(ctx, "event_publisher", string(msg)); err != nil {
				h.log.Error("websocket_handler: failed to publish message to pubsub broker", err)
				return
			}
		}
	}()

	select {}
}
