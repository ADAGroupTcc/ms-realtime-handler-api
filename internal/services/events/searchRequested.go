package events

import (
	"context"
	"encoding/json"
	"fmt"

	sorterApi "github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/clients/sorter"
	"github.com/ADAGroupTcc/ms-realtime-handler-api/internal/http/domain"
)

type SearchRequested struct {
	sorterApi sorterApi.SorterApi
}

func NewSearchRequested(sorterApi sorterApi.SorterApi) Services {
	return &SearchRequested{sorterApi}
}

func (s *SearchRequested) Handle(ctx context.Context, eventToParse []byte) []*domain.EventToPublish {
	var events = make([]*domain.EventToPublish, 0)
	eventInit := domain.SearchRequested{}

	err := json.Unmarshal(eventToParse, &eventInit)
	if err != nil {
		fmt.Println("Error parsing event to publish")
		return events
	}

	sortResponse, err := s.sorterApi.Sort(ctx, eventInit.UserId)
	if err != nil {
		fmt.Println("Error sorting, ", err)
		return events
	}

	users := sortResponse.Users

	for _, user := range users {
		event := domain.EventToPublish{
			UserId: user.Id,
			Event:  domain.CHANNEL_FOUND,
		}
		events = append(events, &event)
	}

	return events
}
