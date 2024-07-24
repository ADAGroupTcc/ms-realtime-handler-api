package services

import (
	"context"

	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/http/domain"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/util"
)

type SubscribeServicer interface {
	SubscribeAsync(ctx context.Context, log *logger.Logger) error
	HandleSubscriptionResponse(podName string, log *logger.Logger)
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

func (s *SubscribeEventService) SubscribeAsync(ctx context.Context, log *logger.Logger) error {
	s.broker.Subscriber.SubscribeAsync(ctx, s.topic, s.subscribeChan)
	return nil
}

func (s *SubscribeEventService) HandleSubscriptionResponse(podName string, log *logger.Logger) {
	for subscribedEvent := range s.subscribeChan {
		subscribedEvent, err := domain.ParseEventToSendToReceiver(subscribedEvent)
		if err != nil {
			log.Error(util.UnableToParseEventResponse, err)
			continue
		}

		responseConn := s.wsConnectionService.GetConn(subscribedEvent.UserId).Conn
		if responseConn == nil {
			log.Infof(util.ReceiverNotOnlineInPod, subscribedEvent.UserId, podName)
			continue
		}
		responseConn.WriteJSON(subscribedEvent)
		log.Infof("websocket_handler: event %s sent to receiver_id %s", subscribedEvent.EventType, subscribedEvent.UserId)
	}
}
