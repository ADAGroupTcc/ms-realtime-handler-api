package sessionClient

import (
	"context"
	"errors"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/http"
)

type SessionClienter interface {
	ValidateMobileToken(ctx context.Context, token string, correlationId string, log *logger.Logger) (string, error)
}

type sessionClient struct {
	http       http.HttpClienter
	instrument interfaces.Instrument
}

func NewSessionClient(httpClient http.HttpClienter, instrument interfaces.Instrument) SessionClienter {
	return &sessionClient{
		httpClient,
		instrument,
	}
}

func (c *sessionClient) ValidateMobileToken(ctx context.Context, token string, correlationId string, log *logger.Logger) (string, error) {
	resp, err := c.http.Get(ctx, http.ClientConfig{
		Headers:   map[string]string{"Token": token, "X-Request-Id": correlationId},
		Endpoint:  "/tokens",
		MetricUrl: "/tokens",
	})

	if err != nil {
		log.Error("session_client: error on call http client api", err)
		return "", err
	}

	if resp.StatusCode == 403 {
		log.Error("session_client: invalid token", errors.New("invalid token"))
		return "", errors.New("invalid token")
	}

	if resp.StatusCode != 200 {
		log.Error("session_client: unable to validate token", errors.New("internal server error"))
		return "", errors.New("error validating token")
	}

	log.Info("session_client: token validated")
	return string(resp.Body), nil
}
