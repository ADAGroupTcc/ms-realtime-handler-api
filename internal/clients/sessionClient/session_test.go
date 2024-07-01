package sessionClient

import (
	"context"
	"errors"
	"testing"

	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/mocks"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateWithSucess(t *testing.T) {
	token := "token"
	correlationId := "correlationId"

	response := http.HttpResponse{
		StatusCode: 200,
		Body:       []byte("C123"),
	}
	httpMock := &mocks.HttpClienter{}
	httpMock.On("Get", mock.Anything, mock.Anything).
		Return(&response, nil).
		Once()

	c := context.Background()

	log := logger.New()

	session := NewSessionClient(httpMock, nil)
	sessionAPIResponse, err := session.ValidateMobileToken(c, token, correlationId, log)

	assert.Nil(t, err)
	assert.Equal(t, sessionAPIResponse, "C123")
}

func TestValidateWithInvalidToken(t *testing.T) {
	token := "invalidToken"
	correlationId := "correlationId"

	response := http.HttpResponse{
		StatusCode: 403,
		Body:       []byte("unauthorized"),
	}
	httpMock := &mocks.HttpClienter{}
	httpMock.On("Get", mock.Anything, mock.Anything).
		Return(&response, nil).
		Once()

	c := context.Background()

	log := logger.New()

	session := NewSessionClient(httpMock, nil)
	sessionAPIResponse, err := session.ValidateMobileToken(c, token, correlationId, log)

	assert.Equal(t, err.Error(), "invalid token")
	assert.Equal(t, sessionAPIResponse, "")
}

func TestValidateWithInternalServerError(t *testing.T) {
	token := "token"
	correlationId := "correlationId"

	response := http.HttpResponse{
		StatusCode: 500,
		Body:       []byte("internal server error"),
	}
	httpMock := &mocks.HttpClienter{}
	httpMock.On("Get", mock.Anything, mock.Anything).
		Return(&response, nil).
		Once()

	c := context.Background()

	log := logger.New()

	session := NewSessionClient(httpMock, nil)
	sessionAPIResponse, err := session.ValidateMobileToken(c, token, correlationId, log)

	assert.Equal(t, err.Error(), "error validating token")
	assert.Equal(t, sessionAPIResponse, "")
}

func TestValidateWithError(t *testing.T) {
	token := "token"
	correlationId := "correlationId"

	httpMock := &mocks.HttpClienter{}
	httpMock.On("Get", mock.Anything, mock.Anything).
		Return(nil, errors.New("some error")).
		Once()

	c := context.Background()

	log := logger.New()

	session := NewSessionClient(httpMock, nil)
	sessionAPIResponse, err := session.ValidateMobileToken(c, token, correlationId, log)

	assert.Equal(t, err.Error(), "some error")
	assert.Equal(t, sessionAPIResponse, "")
}
