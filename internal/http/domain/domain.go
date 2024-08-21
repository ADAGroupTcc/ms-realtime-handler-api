package domain

import (
	"encoding/json"
	"errors"
)

type EventReceived struct {
	EventType string      `json:"event_type"`
	EventId   string      `json:"event_id"`
	Data      interface{} `json:"data"`
}

func (eventReceived *EventReceived) Validate() error {
	if eventReceived.EventType == "" {
		return errors.New("event_type is required")
	}

	if eventReceived.EventId == "" {
		return errors.New("event_id is required")
	}

	if eventReceived.Data == nil {
		return errors.New("data is required")
	}

	return nil
}

func (EventReceived *EventReceived) ToEventToPublish(userId string) *EventToPublish {
	return &EventToPublish{
		Event:   EventReceived.EventType,
		EventId: EventReceived.EventId,
		UserId:  userId,
		Data:    EventReceived.Data,
	}
}

type EventToPublish struct {
	Event   string      `json:"event"`
	EventId string      `json:"event_id"`
	UserId  string      `json:"user_id"`
	Data    interface{} `json:"data"`
}

type EventSubscribed struct {
	Event   string      `json:"event"`
	EventId string      `json:"event_id"`
	UserId  string      `json:"user_id"`
	Data    interface{} `json:"data"`
}

type WsEventResponse struct {
	EventType string `json:"event_type"`
	EventId   string `json:"event_id"`
	Data      string `json:"data"`
}

func ParseEventToSendToReceiver(event []byte) (*EventSubscribed, error) {
	eventSubscribedDomain := EventSubscribed{}
	err := json.Unmarshal(event, &eventSubscribedDomain)
	if err != nil {
		return nil, err
	}
	return &eventSubscribedDomain, nil
}

func ParseEventToWsResponse(event *EventSubscribed) (*WsEventResponse, error) {
	jsonData, err := json.Marshal(event.Data)
	if err != nil {
		return nil, err
	}
	eventSubscribedDomain := WsEventResponse{EventType: event.Event, EventId: event.EventId, Data: string(jsonData)}

	return &eventSubscribedDomain, nil
}
