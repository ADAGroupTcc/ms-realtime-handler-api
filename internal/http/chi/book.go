package chi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book"
	api "github.com/PicPay/lib-go-api"
	"github.com/PicPay/lib-go-instrumentation/interfaces"
	pperr "github.com/PicPay/lib-go-pperr"
	"github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
)

type bookResponseData struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Pages    int    `json:"pages"`
	Quantity int    `json:"quantity"`
}

type bookRequest struct {
	Title    string `json:"title" validate:"required"`
	Author   string `json:"author" validate:"required"`
	Pages    int    `json:"pages" validate:"numeric,gte=1"`
	Quantity int    `json:"quantity" validate:"required,numeric"`
}

// Validate book data
func (r *bookRequest) Validate() error {
	return validate.Struct(r)
}

// Sanitize book data
func (r *bookRequest) Sanitize() {
	p := bluemonday.UGCPolicy()
	r.Title = strings.TrimSpace(p.Sanitize(r.Title))
	r.Author = strings.TrimSpace(p.Sanitize(r.Author))
}

// Interface guard - para bookRequest que o userRequest implementa a interface RequestData
var (
	_ api.RequestData = (*bookRequest)(nil)
)

func getBook(service book.UseCase, instrument interfaces.Instrument) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := instrument.StartContextTransaction(r.Context(), "getBook")
		defer cancel()
		instrument.AddAttributes(ctx, map[string]any{
			"attribute01": "value",
			"attribute02": true,
			"attribute03": 1,
		})
		// validar que o ID é um inteiro e não algo diferente que possa ser usado pra ataque de segurança
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, http.StatusBadRequest)
			return
		}
		b, err := service.Get(ctx, id)
		if err != nil {
			st := http.StatusInternalServerError
			if pperr.Code(err) == pperr.ENOTFOUND {
				st = http.StatusNotFound
			}
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, st)
			return
		}
		rd := bookResponseData{
			ID:       b.ID,
			Title:    b.Title,
			Author:   b.Author,
			Pages:    b.Pages,
			Quantity: b.Quantity,
		}
		out := api.Response[bookResponseData]{
			Data: rd,
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(out); err != nil {
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, http.StatusInternalServerError)
			return
		}
	})
}

func createBook(service book.UseCase, instrument interfaces.Instrument) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := instrument.StartContextTransaction(r.Context(), "createBook")
		defer cancel()
		data := bookRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		err = data.Validate()
		if err != nil {
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, http.StatusBadRequest)
			return
		}
		data.Sanitize()
		b, err := service.Create(ctx, data.Title, data.Author, data.Pages, data.Quantity)
		if err != nil {
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		rd := bookResponseData{
			ID:       b.ID,
			Title:    b.Title,
			Author:   b.Author,
			Pages:    b.Pages,
			Quantity: b.Quantity,
		}
		out := api.Response[bookResponseData]{
			Data: rd,
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(out); err != nil {
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, http.StatusInternalServerError)
			return
		}
	})
}
