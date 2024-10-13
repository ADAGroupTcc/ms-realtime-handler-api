package services

import (
	"context"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/pubsubconnector"
)

type SubscribeServicer interface {
	SubscribeAsync(ctx context.Context) error
	HandleSubscriptionResponse()
}

type SubscribeEventService struct {
	broker              *pubsubconnector.PubSubBroker
	subscribeChan       chan []byte
	topic               string
	wsConnectionService WsConnectionServicer
}

func NewSubscribeEventService(broker *pubsubconnector.PubSubBroker, topicToSubscribe string, subscribeChan chan []byte, wsConnectionService WsConnectionServicer) SubscribeServicer {
	return &SubscribeEventService{
		broker:              broker,
		subscribeChan:       subscribeChan,
		topic:               topicToSubscribe,
		wsConnectionService: wsConnectionService,
	}
}

func (s *SubscribeEventService) SubscribeAsync(ctx context.Context) error {
	s.broker.Subscriber.SubscribeAsync(ctx, s.topic, s.subscribeChan)
	return nil
}

func (s *SubscribeEventService) HandleSubscriptionResponse() {
	for subscribedEvent := range s.subscribeChan {
		eventReceived, err := domain.ParseEventToSendToReceiver(subscribedEvent)
		if err != nil {
			continue
		}
		activeConn := s.wsConnectionService.GetConn(eventReceived.UserId)
		if activeConn == nil {
			continue
		}
		eventToPublish, err := domain.ParseEventToWsResponse(eventReceived)
		if err != nil {
			continue
		}
		activeConn.Conn.WriteJSON(eventToPublish)
	}
}
