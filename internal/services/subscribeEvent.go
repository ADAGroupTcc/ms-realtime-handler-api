package services

import (
	"context"

	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
)

type SubscribeServicer interface {
	SubscribeAsync(ctx context.Context, eventsChan chan []byte, log *logger.Logger) error
}

type SubscribeEventService struct {
	broker *pubsubconnector.PubSubBroker
	topic  string
}

func NewSubscribeEventService(broker *pubsubconnector.PubSubBroker, topicToSubscribe string) SubscribeServicer {
	return &SubscribeEventService{
		broker: broker,
		topic:  topicToSubscribe,
	}
}

func (s *SubscribeEventService) SubscribeAsync(ctx context.Context, eventsChan chan []byte, log *logger.Logger) error {
	s.broker.Subscriber.SubscribeAsync(ctx, s.topic, eventsChan)
	return nil
}
