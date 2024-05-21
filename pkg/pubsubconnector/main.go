package pubsubconnector

import (
	"context"
)

type Event struct {
	Type  string `json:"type"`
	Event string `json:"event"`
}

type PubSubBroker struct {
	Publisher  Publisher
	Subscriber Subscriber
}

type Publisher interface {
	Publisher(ctx context.Context, channel string, message string) error
}

type Subscriber interface {
	Subscriber(ctx context.Context, channel string) (*Event, error)
}

func NewPubSubBroker(Publisher Publisher, Subscriber Subscriber) *PubSubBroker {
	return &PubSubBroker{
		Publisher:  Publisher,
		Subscriber: Subscriber,
	}
}
