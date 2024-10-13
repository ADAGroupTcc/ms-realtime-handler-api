package services

import (
	"context"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/pkg/pubsubconnector"
)

type PublishServicer interface {
	PublishEvent(ctx context.Context, message interface{}) error
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

func (p *PublishEventService) PublishEvent(ctx context.Context, message interface{}) error {
	if err := p.broker.Publisher.Publish(ctx, message, &map[string]interface{}{
		"topic": p.topic,
	}); err != nil {
		return err
	}

	return nil
}
