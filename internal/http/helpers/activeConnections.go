package helpers

import "github.com/gorilla/websocket"

type activeConnections struct {
	users map[string]*websocket.Conn
}

func (activeConns *activeConnections) SetConn(id string, conn *websocket.Conn) {
	activeConns.users[id] = conn
}

func (activeConns *activeConnections) GetConn(id string) *websocket.Conn {
	return activeConns.users[id]
}

func (activeConns *activeConnections) DeleteConn(id string) {
	connToDelete := activeConns.GetConn(id)
	if connToDelete != nil {
		connToDelete.Close()
	}
	delete(activeConns.users, id)
}

func (activeConns *activeConnections) ConnectionSize() int {
	return len(activeConns.users)
}

func NewActiveConnections() *activeConnections {
	return &activeConnections{
		users: make(map[string]*websocket.Conn),
	}
}
