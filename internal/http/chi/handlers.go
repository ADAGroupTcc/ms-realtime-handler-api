package chi

import (
	"context"
	"net/http"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book"
	pgbook "github.com/PicPay/ms-chatpicpay-websocket-handler-api/book/postgresql"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/user"
	pguser "github.com/PicPay/ms-chatpicpay-websocket-handler-api/user/postgresql"
	api_middleware "github.com/PicPay/lib-go-api/middleware"
	chilog "github.com/PicPay/lib-go-chi-logger"
	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Handlers define http handlers
func Handlers(ctx context.Context, l *logger.Logger, s book.UseCase, instrument interfaces.Instrument) *chi.Mux {
	validate = validator.New(validator.WithRequiredStructEnabled())
	bRepo := pgbook.New()
	bService := book.NewService(bRepo)

	uRepo := pguser.New()
	uService := user.NewService(uRepo)
	r := chi.NewRouter()
	r.Use(chilog.ChiLogger(l))
	r.Use(middleware.Recoverer)
	r.Use(api_middleware.ResponseHeader)
	for _, mdl := range instrument.NewMiddlewaresFactory(ctx).NewDefaultMiddlewaresForChi(ctx) {
		r.Use(mdl)
	}
	r.Get("/health", healthHandler)
	r.Method(http.MethodGet, "/v1/books/{id}", getBook(bService, instrument))
	r.Method(http.MethodPost, "/v1/books", createBook(bService, instrument))

	r.Method(http.MethodGet, "/v1/users/{id}", getUser(uService, instrument))
	r.Method(http.MethodPost, "/v1/users", createUser(uService, instrument))
	return r
}
