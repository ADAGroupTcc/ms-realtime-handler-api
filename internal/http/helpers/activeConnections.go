package helpers

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type activeConnections struct {
	users    map[string]*websocket.Conn
	connTime map[string]time.Time
	mutex    sync.RWMutex
}

func (activeConns *activeConnections) SetConn(id string, conn *websocket.Conn) {
	activeConns.mutex.Lock()
	activeConns.users[id] = conn
	activeConns.connTime[id] = time.Now()
	activeConns.mutex.Unlock()
}

func (activeConns *activeConnections) GetConn(id string) *websocket.Conn {
	activeConns.mutex.RLock()
	con := activeConns.users[id]
	activeConns.mutex.RUnlock()
	return con
}

func (activeConns *activeConnections) DeleteConn(id string) {
	connToDelete := activeConns.GetConn(id)
	if connToDelete != nil {
		connToDelete.Close()
	}
	activeConns.mutex.Lock()
	delete(activeConns.users, id)
	activeConns.mutex.Unlock()
}

func (activeConns *activeConnections) ConnectionSize() int {
	activeConns.mutex.RLock()
	amount := len(activeConns.users)
	activeConns.mutex.RUnlock()
	return amount
}

func (activeConns *activeConnections) GetConnTime(id string) time.Time {
	activeConns.mutex.RLock()
	connTime := activeConns.connTime[id]
	activeConns.mutex.RUnlock()
	return connTime
}

func NewActiveConnections() *activeConnections {
	return &activeConnections{
		users:    make(map[string]*websocket.Conn),
		connTime: make(map[string]time.Time),
	}
}
