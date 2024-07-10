package helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestCreateNewMapOfActiveConnectionsWithSucess(t *testing.T) {
	activeConnections := NewActiveConnections()

	assert.NotNil(t, activeConnections)
	assert.Equal(t, 0, activeConnections.ConnectionSize())
}

func TestAddNewConnectionInActiveConnectionsMap(t *testing.T) {
	activeConnections := NewActiveConnections()

	wsConn := websocket.Conn{}
	activeConnections.SetConn("C123", &wsConn)

	assert.NotNil(t, activeConnections)
	assert.Equal(t, 1, activeConnections.ConnectionSize())
	assert.Equal(t, &wsConn, activeConnections.GetConn("C123"))
}

func TestGetAllActiveConnections(t *testing.T) {
	activeConnections := NewActiveConnections()
	wsConn := websocket.Conn{}

	activeConnections.SetConn("C123", &wsConn)
	activeConnections.SetConn("C456", &wsConn)
	activeConnections.SetConn("C789", &wsConn)

	assert.NotNil(t, activeConnections)
	assert.Equal(t, 3, activeConnections.ConnectionSize())
}

func TestGetAnUnknowConnection(t *testing.T) {
	activeConnections := NewActiveConnections()
	wsConn := websocket.Conn{}

	activeConnections.SetConn("C123", &wsConn)
	activeConnections.SetConn("C456", &wsConn)
	activeConnections.SetConn("C789", &wsConn)

	assert.NotNil(t, activeConnections)
	assert.Equal(t, 3, activeConnections.ConnectionSize())
	assert.Nil(t, activeConnections.GetConn("UNKNOWN"))
}

func TestGetaAnUnknowConnection(t *testing.T) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	ws, _ := upgrader.Upgrade(res, req, http.Header{})
	activeConnections := NewActiveConnections()

	activeConnections.SetConn("C123", ws)
	activeConnections.SetConn("C456", ws)
	activeConnections.SetConn("C789", ws)

	activeConnections.DeleteConn("C456")

	assert.NotNil(t, activeConnections)
	assert.Equal(t, 2, activeConnections.ConnectionSize())
	assert.Nil(t, activeConnections.GetConn("C456"))
}
