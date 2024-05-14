package chi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book/postgresql"
	"github.com/PicPay/lib-go-instrumentation/instruments/dummy"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	l := logger.New()
	ctx := context.Background()
	s := book.NewService(postgresql.New())
	instrument := dummy.NewInstrument()
	h := Handlers(context.TODO(), l, s, instrument)
	t.Run("ShouldReturnOK", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/health", nil)
		assert.Nil(t, err)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
}
