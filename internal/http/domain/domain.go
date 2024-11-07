package domain

import (
	"encoding/json"
	"errors"
	"time"
)

type EventReceived struct {
	EventType string      `json:"event"`
	EventId   string      `json:"event_id"`
	Data      interface{} `json:"data"`
}

func (eventReceived *EventReceived) Validate() error {
	if eventReceived.EventType == "" {
		return errors.New("event is required")
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
	EventType string `json:"event"`
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

type MessageSent struct {
	Event   string          `json:"event"`
	EventId string          `json:"event_id"`
	UserId  string          `json:"user_id"`
	Data    MessageReceived `json:"data"`
}

func (m *MessageSent) ParseMessageSentToEventToPublish(messageCreated *MessageCreated) *EventToPublish {
	return &EventToPublish{
		Event:   m.Event,
		EventId: m.EventId,
		UserId:  m.UserId,
		Data:    messageCreated,
	}
}

func (m *MessageSent) ParseMessageSentToMessageRequest() MessageRequest {
	return MessageRequest{
		ChannelId: m.Data.Channel.ChannelId,
		SenderId:  m.UserId,
		Message:   m.Data.Message,
		Data:      m.Data.Data,
	}
}

type MessageReceived struct {
	Channel *Channel `json:"channel"`
	Message string   `json:"message"`
	Data    string   `json:"data"`
}

type Channel struct {
	ChannelId string   `json:"channel_id"`
	Members   []string `json:"members"`
}

type MessageRequest struct {
	ChannelId string `param:"channel_id"`
	SenderId  string `json:"sender_id"`
	Message   string `json:"message"`
	Data      string `json:"data"`
}

type MessageCreated struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ChannelId string    `json:"channel_id"`
	SenderId  string    `json:"sender_id"`
	Message   string    `json:"message"`
	Data      string    `json:"data"`
	IsEdited  bool      `json:"is_edited"`
}

func (messageCreated *MessageCreated) ParseMessageCreatedToEventToPublish(receiverId string, eventId string) *EventToPublish {
	return &EventToPublish{
		Event:   "MESSAGE_RECEIVED",
		EventId: eventId,
		UserId:  receiverId,
		Data:    messageCreated,
	}
}

type SortResponse struct {
	Users      []User   `json:"users"`
	Categories []string `json:"categories"`
}

type User struct {
	Id        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
}

type SearchRequested struct {
	Event  string `json:"event"`
	UserId string `json:"user_id"`
}

type ChannelEvents struct {
	Event string   `json:"event"`
	Users []string `json:"users"`
}

const (
	SEARCH_REQUESTED = "SEARCH_REQUESTED"
	CHANNEL_ACCEPTED = "CHANNEL_ACCEPTED"
	CHANNEL_REJECTED = "CHANNEL_REJECTED"
	CHANNEL_FOUND    = "CHANNEL_FOUND"
)

/*
// entrada
{
event: SEARCH_REQUESTED
user_id: "123"
}

// saida
{
event: CHANNEL_FOUND,
data: {
	"users": len(4),
	"categories": ["category1", "category2"]
}
}

1..*
{
event: CHANNEL_ACCEPTED | CHANNEL_REJECTED
user_id: "123"
}

*/
