package chi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/user"
	api "github.com/PicPay/lib-go-api"
	"github.com/PicPay/lib-go-instrumentation/interfaces"
	pperr "github.com/PicPay/lib-go-pperr"
	"github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
)

type userResponseData struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Validate user data
func (r *userRequest) Validate() error {
	return validate.Struct(r)
}

// Sanitize user data
func (r *userRequest) Sanitize() {
	p := bluemonday.UGCPolicy()
	r.Name = strings.TrimSpace(p.Sanitize(r.Name))
	r.Email = strings.TrimSpace(p.Sanitize(r.Email))
}

// Interface guard - para garantir que o userRequest implementa a interface RequestData
var (
	_ api.RequestData = (*userRequest)(nil)
)

func getUser(service user.UseCase, instrument interfaces.Instrument) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := instrument.StartContextTransaction(r.Context(), "getUser")
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
		u, err := service.Get(ctx, id)
		if err != nil {
			st := http.StatusInternalServerError
			if pperr.Code(err) == pperr.ENOTFOUND {
				st = http.StatusNotFound
			}
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, st)
			return
		}
		rd := userResponseData{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		}
		out := api.Response[userResponseData]{
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

func createUser(service user.UseCase, instrument interfaces.Instrument) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := instrument.StartContextTransaction(r.Context(), "createUser")
		defer cancel()
		data := userRequest{}
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
		u, err := service.Create(ctx, data.Name, data.Email, data.Password)
		if err != nil {
			instrument.NoticeError(ctx, err)
			api.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		rd := userResponseData{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		}
		out := api.Response[userResponseData]{
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
