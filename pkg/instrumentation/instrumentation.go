package instrumentation

import (
	"context"
	"time"

	"github.com/PicPay/lib-go-instrumentation/instruments/opentelemetry"
	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	pperr "github.com/PicPay/lib-go-pperr"
)

type Config struct {
	Context             context.Context
	AppName             string
	AppEnv              string
	TraceEndpoint       string
	MetricEndpoint      string
	SendMetricsInterval time.Duration
	Logger              *logger.Logger
}

func New(config Config) interfaces.Instrument {
	instrumentation, err := opentelemetry.NewOtelInstrument(
		config.Context,
		&opentelemetry.Config{
			AppName:             config.AppName,
			AppEnv:              config.AppEnv,
			TraceEndpoint:       config.TraceEndpoint,
			MetricEndpoint:      config.MetricEndpoint,
			Logger:              config.Logger,
			SendMetricsInterval: config.SendMetricsInterval,
		},
	)

	if err != nil {
		config.Logger.Panic("error to initialize instrumentation", pperr.Wrap(err))
	}

	return instrumentation
}
