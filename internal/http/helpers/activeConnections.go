package helpers

import (
	"github.com/gorilla/websocket"
	"sync"
)

type activeConnections struct {
	users map[string]*websocket.Conn
	mutex sync.RWMutex
}

func (activeConns *activeConnections) SetConn(id string, conn *websocket.Conn) {
	activeConns.mutex.Lock()
	activeConns.users[id] = conn
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

func NewActiveConnections() *activeConnections {
	return &activeConnections{
		users: make(map[string]*websocket.Conn),
	}
}
