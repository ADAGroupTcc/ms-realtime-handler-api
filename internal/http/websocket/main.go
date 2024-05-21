package websocket

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	users = map[string]*websocket.Conn{}
)

type websocketHandler struct {
	broker     *pubsubconnector.PubSubBroker
	instrument interfaces.Instrument
}

func NewHandler(broker *pubsubconnector.PubSubBroker, instrument interfaces.Instrument) *websocketHandler {
	return &websocketHandler{
		broker:     broker,
		instrument: instrument,
	}
}

func (h *websocketHandler) WebsocketServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade connection:", err)
			return
		}
		defer conn.Close()

		fmt.Println("Conectado: ", r.Header.Get("user_id"))

		user_id := r.Header.Get("user_id")
		users[user_id] = conn

		fmt.Println("Conexões: ", users)
		ctx := r.Context()

		go func() {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					if strings.Contains(err.Error(), "websocket: close 1000") {
						fmt.Println("Conexão encerrada: ", r.Header.Get("user_id"))
						delete(users, r.Header.Get("user_id"))
						fmt.Println("Conexões: ", users)
						return
					}
					log.Println("Failed to read message from WebSocket client:", err)
					return
				}

				if err := h.broker.Publisher.Publisher(ctx, "event_publisher", string(msg)); err != nil {
					log.Println("Failed to publish message to Redis PubSub:", err)
					return
				}
			}
		}()

		select {}
	}
}
