package chi

import (
	"context"
	"testing"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book/postgresql"
	"github.com/PicPay/lib-go-instrumentation/instruments/dummy"
	"github.com/stretchr/testify/assert"
)

func TestHandlers(t *testing.T) {
	s := book.NewService(postgresql.New())
	instrument := dummy.NewInstrument()
	h := Handlers(context.TODO(), nil, s, instrument)
	routes := h.Routes()
	assert.Equal(t, 5, len(routes))
	assert.Equal(t, "/health", routes[0].Pattern)
	assert.Equal(t, "/v1/books", routes[1].Pattern)
	assert.Equal(t, "/v1/books/{id}", routes[2].Pattern)

	assert.Equal(t, "/v1/users", routes[3].Pattern)
	assert.Equal(t, "/v1/users/{id}", routes[4].Pattern)
}
