package chi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book/postgresql"
	api "github.com/PicPay/lib-go-api"
	"github.com/PicPay/lib-go-instrumentation/instruments/dummy"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetBook(t *testing.T) {
	l := logger.New()
	ctx := context.Background()
	s := book.NewService(postgresql.New())
	instrument := dummy.NewInstrument()
	h := Handlers(context.TODO(), l, s, instrument)
	t.Run("Should Return OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/v1/books/1", nil)
		assert.Nil(t, err)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		var result api.Response[bookResponseData]
		assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Equal(t, 1, result.Data.ID)
		assert.Equal(t, "Fake Book", result.Data.Title)
	})
	t.Run("Should Return NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/v1/books/2", nil)
		assert.Nil(t, err)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		var result api.HTTPError
		assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Contains(t, result.Detail, "Error reading book")
		assert.Equal(t, "object_not_found", result.Type)
	})
}

func TestCreateBook(t *testing.T) {
	l := logger.New()
	ctx := context.Background()
	s := book.NewService(postgresql.New())
	instrument := dummy.NewInstrument()
	h := Handlers(context.TODO(), l, s, instrument)
	t.Run("Should Return OK", func(t *testing.T) {
		body := `{
			"title": "title of the book",
			"author": "author of the book",
			"pages": 190,
			"quantity": 20
		  }`

		bodyReader := bytes.NewReader([]byte(body))
		w := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/v1/books", bodyReader)
		assert.Nil(t, err)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Result().StatusCode)

		var result api.Response[bookResponseData]
		assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Equal(t, 1, result.Data.ID)
		assert.Equal(t, "title of the book", result.Data.Title)
	})
	t.Run("Should ReturnInvalidRequest with single missing field", func(t *testing.T) {
		body := `{
			"title": "title of the book",
			"author": "author of the book",
			"quantity": 20
		  }`

		bodyReader := bytes.NewReader([]byte(body))
		w := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/v1/books", bodyReader)
		assert.Nil(t, err)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

		var result api.HTTPError
		assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Contains(t, result.Detail, "Field validation for 'Pages' failed")
		assert.Equal(t, "invalid_request", result.Type)
	})
	t.Run("Should ReturnInvalidRequest with multiple missing field", func(t *testing.T) {
		body := `{
			"title": "title of the book",
			"author": "author of the book"
		  }`

		bodyReader := bytes.NewReader([]byte(body))
		w := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/v1/books", bodyReader)
		assert.Nil(t, err)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

		var result api.HTTPError
		assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Contains(t, result.Detail, "Field validation for 'Pages' failed")
		assert.Contains(t, result.Detail, "Field validation for 'Quantity' failed")
		assert.Equal(t, "invalid_request", result.Type)
	})
	t.Run("Should sanitize data", func(t *testing.T) {
		body := `{
  			"title": "<script>alert</script>title of the book    ",
  			"author": "   author of the book",
  			"pages": 1,
  			"quantity": 20
		}`

		bodyReader := bytes.NewReader([]byte(body))
		w := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/v1/books", bodyReader)
		assert.Nil(t, err)
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Result().StatusCode)

		var result api.Response[bookResponseData]
		assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Equal(t, 1, result.Data.ID)
		assert.Equal(t, "title of the book", result.Data.Title)
		assert.Equal(t, "author of the book", result.Data.Author)
	})
}
