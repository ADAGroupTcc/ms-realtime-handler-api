package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
)

type ChannelEvents struct {
	event string
}

func NewChannelEvents(event string) Services {
	return &ChannelEvents{event}
}

func (s *ChannelEvents) Handle(ctx context.Context, eventToParse []byte) []*domain.EventToPublish {
	var events = make([]*domain.EventToPublish, 0)
	eventInit := domain.ChannelEvents{}

	err := json.Unmarshal(eventToParse, &eventInit)
	if err != nil {
		fmt.Println("Error parsing event to publish")
		return events
	}

	users := eventInit.Users

	for _, userId := range users {
		event := domain.EventToPublish{
			UserId: userId,
			Event:  s.event,
		}
		events = append(events, &event)
	}

	return events
}
