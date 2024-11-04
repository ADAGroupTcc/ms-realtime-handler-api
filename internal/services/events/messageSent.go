package events

import (
	"context"
	"encoding/json"
	"fmt"

	messagesClient "github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/clients/messages"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
)

type Services interface {
	Handle(ctx context.Context, eventToParse []byte) []*domain.EventToPublish
}

type MessageSent struct {
	messagesApi messagesClient.MessagesApi
}

func NewMessageSent(messagesClient messagesClient.MessagesApi) Services {
	return &MessageSent{messagesClient}
}

func (s *MessageSent) Handle(ctx context.Context, eventToParse []byte) []*domain.EventToPublish {
	var events = make([]*domain.EventToPublish, 0)
	eventInit := domain.MessageSent{}

	err := json.Unmarshal(eventToParse, &eventInit)
	if err != nil {
		fmt.Println("Erro ao fazer unmarshal do evento recebido:", err)
		return events
	}

	messageRequest := domain.MessageRequest{
		ChannelId: eventInit.Data.Channel.ChannelId,
		SenderId:  eventInit.UserId,
		Message:   eventInit.Data.Message,
		Data:      eventInit.Data.Data,
	}

	headers := map[string]string{
		"X-Request-Id": eventInit.EventId,
	}

	messageCreated, err := s.messagesApi.CreateMessage(ctx, messageRequest, headers)
	if err != nil {
		fmt.Println("Erro ao criar mensagem:", err)
		return events
	}

	members := eventInit.Data.Channel.Members
	for _, memberId := range members {
		if memberId != messageCreated.SenderId {
			event := messageCreated.ParseMessageCreatedToEventToPublish(memberId, eventInit.EventId)
			events = append(events, event)
			continue
		}

		event := &domain.EventToPublish{
			Event:   "MESSAGE_CREATED",
			EventId: eventInit.EventId,
			UserId:  eventInit.UserId,
			Data:    messageCreated,
		}
		events = append(events, event)
	}

	return events
}
