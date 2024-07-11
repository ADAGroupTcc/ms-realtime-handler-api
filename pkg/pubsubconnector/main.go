package pubsubconnector

import (
	"context"
)

type Event struct {
	EventName string      `json:"eventName"`
	Event     interface{} `json:"event"`
}

type PubSubBroker struct {
	Publisher  Publisher
	Subscriber Subscriber
}

type Publisher interface {
	Publish(ctx context.Context, message interface{}, configMap *map[string]interface{}) error
}

type Subscriber interface {
	SubscribeAsync(ctx context.Context, topic string, eventsChan chan []byte)
}

func NewPubSubBroker(Publisher Publisher, Subscriber Subscriber) *PubSubBroker {
	return &PubSubBroker{
		Publisher:  Publisher,
		Subscriber: Subscriber,
	}
}
