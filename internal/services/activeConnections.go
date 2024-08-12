package services

import (
	"context"
	"os"
	"sync"
	"time"

	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/cache"
	"github.com/gorilla/websocket"
)

var mutex sync.RWMutex
var POD_NAME = os.Getenv("HOSTNAME")

type ActiveConn struct {
	PodName string
	Conn    *websocket.Conn
	Time    time.Time
}

type WsConnectionServicer interface {
	SetConn(ctx context.Context, userId string, conn *websocket.Conn)
	GetConn(userId string) *ActiveConn
	DeleteConn(ctx context.Context, userId string)
	ConnectionSize() int
	GetConnStartTime(userId string) time.Time
	RefreshConnection(ctx context.Context, userId string)
}

type websocketConnections struct {
	actives          map[string]*ActiveConn
	readDeadlineWait time.Duration
	cache            cache.Cache
	log              *logger.Logger
}

func NewWebsocketConnectionsService(readDeadlineWait time.Duration, cache cache.Cache, log *logger.Logger) *websocketConnections {
	return &websocketConnections{
		actives:          make(map[string]*ActiveConn),
		readDeadlineWait: readDeadlineWait,
		cache:            cache,
		log:              log,
	}
}

func (wsConnection *websocketConnections) SetConn(ctx context.Context, userId string, conn *websocket.Conn) {
	mutex.Lock()
	wsConnection.actives[userId] = &ActiveConn{
		PodName: os.Getenv("HOSTNAME"),
		Conn:    conn,
		Time:    time.Now(),
	}
	mutex.Unlock()
	wsConnection.cache.Set(ctx, userId, POD_NAME)
}

func (wsConnection *websocketConnections) GetConn(userId string) *ActiveConn {
	mutex.RLock()
	conn := wsConnection.actives[userId]
	mutex.RUnlock()
	return conn
}

func (wsConnection *websocketConnections) DeleteConn(ctx context.Context, userId string) {
	connToDelete := wsConnection.GetConn(userId)
	if connToDelete != nil {
		connToDelete.Conn.Close()
	}
	mutex.Lock()
	delete(wsConnection.actives, userId)
	mutex.Unlock()
	wsConnection.cache.Delete(ctx, userId)
}

func (wsConnection *websocketConnections) ConnectionSize() int {
	mutex.RLock()
	amount := len(wsConnection.actives)
	mutex.RUnlock()
	return amount
}

func (wsConnection *websocketConnections) GetConnStartTime(userId string) time.Time {
	mutex.RLock()
	connTime := wsConnection.actives[userId].Time
	mutex.RUnlock()
	return connTime
}

func (wsConnection *websocketConnections) RefreshConnection(ctx context.Context, userId string) {
	userConn := wsConnection.GetConn(userId)
	if userConn == nil {
		return
	}

	conn := userConn.Conn

	conn.SetReadDeadline(time.Now().Add(wsConnection.readDeadlineWait))

	conn.SetPongHandler(func(appData string) error {
		wsConnection.log.Debugf("pong received from user: %s", userId)
		wsConnection.cache.Set(ctx, userId, POD_NAME)
		conn.SetReadDeadline(time.Now().Add(wsConnection.readDeadlineWait))
		return nil
	})

	ticker := time.NewTicker(wsConnection.readDeadlineWait + 2*time.Second)

	for {
		select {
		case <-ctx.Done():
			wsConnection.log.Debugf("context canceled for user %s", userId)
			return
		case <-ticker.C:
			mutex.Lock()
			err := conn.WriteMessage(websocket.PingMessage, nil)
			mutex.Unlock()

			if err != nil {
				wsConnection.DeleteConn(ctx, userId)
				wsConnection.log.Errorf("could not send ping message to user %s. reason: %s", userId, err.Error())
				ticker.Stop()
				return
			}
			wsConnection.log.Debugf("ping sent to user %s", userId)
		}
	}
}
