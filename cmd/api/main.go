package main

import (
	"context"

	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/book/postgresql"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/config"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/chi" //@todo mude para o framework que for usar. Confira a documentação em https://github.com/PicPay/lib-go-api
	api "github.com/PicPay/lib-go-api"
	"github.com/PicPay/lib-go-instrumentation/instruments/factory"
	logger "github.com/PicPay/lib-go-logger/v2"
	pperr "github.com/PicPay/lib-go-pperr"
	_ "go.uber.org/automaxprocs"
)

func main() {
	l := logger.New(
		logger.WithFatalHook(logger.WriteThenFatal),
	)
	// inicializando variaveis de ambiente
	envs := config.LoadEnvVars(l)
	ctx := context.Background()
	instrument, err := factory.NewInstrumentationFactory(&factory.Config{}).NewInstrument(ctx)
	if err != nil {
		l.Panic("error to initialize instrumentation", pperr.Wrap(err))
	}

	s := book.NewService(postgresql.New())
	//@todo mude para o framework que for usar. Confira a documentação em https://github.com/PicPay/lib-go-api
	h := chi.Handlers(ctx, l, s, instrument)
	err = api.Start(l, envs.APIPort, h)
	if err != nil {
		l.Fatal("error running api", err)
	}
}
