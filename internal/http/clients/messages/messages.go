package messagesClient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/http"
)

type MessagesApi interface {
	CreateMessage(ctx context.Context, messageRequest domain.MessageRequest, headers map[string]string) (*domain.MessageCreated, error)
}

type messagesApi struct {
	messagesClient http.HttpClienter
}

func New(messagesClient http.HttpClienter) MessagesApi {
	return &messagesApi{
		messagesClient: messagesClient,
	}
}

func (m *messagesApi) CreateMessage(ctx context.Context, messageRequest domain.MessageRequest, headers map[string]string) (*domain.MessageCreated, error) {
	requestPayload, err := json.Marshal(messageRequest)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/v1/channels/%s/messages", messageRequest.ChannelId)
	response, err := m.messagesClient.Post(ctx, http.ClientConfig{
		Endpoint: url,
		Headers:  headers,
	}, requestPayload)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 201 {
		return nil, fmt.Errorf("erro ao criar mensagem: %s", response.Body)
	}

	var message *domain.MessageCreated
	err = json.Unmarshal(response.Body, &message)
	if err != nil {
		return nil, err
	}

	return message, nil
}
