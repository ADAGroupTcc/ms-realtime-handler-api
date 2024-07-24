package services

import (
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var mutex sync.RWMutex

type ActiveConn struct {
	PodName string
	Conn    *websocket.Conn
	Time    time.Time
}

type WsConnectionServicer interface {
	SetConn(userId string, conn *websocket.Conn)
	GetConn(userId string) *ActiveConn
	DeleteConn(userId string)
	ConnectionSize() int
	GetConnStartTime(userId string) time.Time
}

type websocketConnections struct {
	actives map[string]*ActiveConn
}

func NewWebsocketConnectionsService() *websocketConnections {
	return &websocketConnections{
		actives: make(map[string]*ActiveConn),
	}
}

func (wsConnection *websocketConnections) SetConn(userId string, conn *websocket.Conn) {
	mutex.Lock()
	wsConnection.actives[userId] = &ActiveConn{
		PodName: os.Getenv("HOSTNAME"),
		Conn:    conn,
		Time:    time.Now(),
	}
	mutex.Unlock()
}

func (wsConnection *websocketConnections) GetConn(userId string) *ActiveConn {
	mutex.RLock()
	conn := wsConnection.actives[userId]
	mutex.RUnlock()
	return conn
}

func (wsConnection *websocketConnections) DeleteConn(userId string) {
	connToDelete := wsConnection.GetConn(userId)
	if connToDelete != nil {
		connToDelete.Conn.Close()
	}
	mutex.Lock()
	delete(wsConnection.actives, userId)
	mutex.Unlock()
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
