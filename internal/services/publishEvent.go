package services

import (
	"context"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/util"

	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/pubsubconnector"
)

type PublishServicer interface {
	PublishEvent(ctx context.Context, message interface{}, log *logger.Logger) error
}

type PublishEventService struct {
	broker *pubsubconnector.PubSubBroker
	topic  string
}

func NewPublishEventService(broker *pubsubconnector.PubSubBroker, topicToPublish string) PublishServicer {
	return &PublishEventService{
		broker: broker,
		topic:  topicToPublish,
	}
}

func (p *PublishEventService) PublishEvent(ctx context.Context, message interface{}, log *logger.Logger) error {
	if err := p.broker.Publisher.Publish(ctx, message, &map[string]interface{}{
		"topic": p.topic,
	}); err != nil {
		log.Error(util.FailedToPublishMessageToPubSubBroker, err)
		return err
	}

	return nil
}
