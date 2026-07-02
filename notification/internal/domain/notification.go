package domain

import (
	"errors"
	"time"
)

type ChannelType string

const (
	ChannelEmail    ChannelType = "email"
	ChannelSMS      ChannelType = "sms"
	ChannelWhatsApp ChannelType = "whatsapp"
)

type Channel struct {
	Type  ChannelType
	Phone string
	Email string
}

type ChannelMetadata struct {
	ProviderResourceID string
}

type Notification struct {
	ID              string
	Title           string
	Content         string
	IsSent          bool
	IsSentConfirmed bool
	Channel         Channel
	ChannelMetadata *ChannelMetadata
	CreatedAt       time.Time
	UpdatedAt       time.Time
	events          []Event
}

type Event interface {
	EventName() string
	NotificationID() string
}

type NotificationCreated struct{ ID string }
type NotificationSent struct{ ID string }
type NotificationSentConfirmed struct{ ID string }

func (event NotificationCreated) EventName() string            { return "NotificationCreated" }
func (event NotificationCreated) NotificationID() string       { return event.ID }
func (event NotificationSent) EventName() string               { return "NotificationSent" }
func (event NotificationSent) NotificationID() string          { return event.ID }
func (event NotificationSentConfirmed) EventName() string      { return "NotificationSentConfirmed" }
func (event NotificationSentConfirmed) NotificationID() string { return event.ID }

func NewNotification(id, title, content string, channel Channel, now time.Time) (Notification, error) {
	if id == "" || title == "" || content == "" {
		return Notification{}, errors.New("id, title and content are required")
	}
	if err := channel.Validate(); err != nil {
		return Notification{}, err
	}
	notification := Notification{
		ID:        id,
		Title:     title,
		Content:   content,
		Channel:   channel,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}
	notification.record(NotificationCreated{ID: id})
	return notification, nil
}

func HydrateNotification(id, title, content string, isSent, isSentConfirmed bool, channel Channel, metadata *ChannelMetadata, createdAt, updatedAt time.Time) Notification {
	return Notification{
		ID:              id,
		Title:           title,
		Content:         content,
		IsSent:          isSent,
		IsSentConfirmed: isSentConfirmed,
		Channel:         channel,
		ChannelMetadata: metadata,
		CreatedAt:       createdAt.UTC(),
		UpdatedAt:       updatedAt.UTC(),
	}
}

func (notification *Notification) MarkSent(metadata ChannelMetadata, now time.Time) {
	if notification.IsSent {
		return
	}
	notification.IsSent = true
	notification.ChannelMetadata = &metadata
	notification.UpdatedAt = now.UTC()
	notification.record(NotificationSent{ID: notification.ID})
}

func (notification *Notification) ConfirmSent(now time.Time) {
	if notification.IsSentConfirmed {
		return
	}
	notification.IsSentConfirmed = true
	notification.UpdatedAt = now.UTC()
	notification.record(NotificationSentConfirmed{ID: notification.ID})
}

func (notification *Notification) PullEvents() []Event {
	events := notification.events
	notification.events = nil
	return events
}

func (notification *Notification) record(event Event) {
	notification.events = append(notification.events, event)
}

func (channel Channel) Validate() error {
	switch channel.Type {
	case ChannelSMS, ChannelWhatsApp:
		if channel.Phone == "" {
			return errors.New("phone is required")
		}
	case ChannelEmail:
		if channel.Email == "" {
			return errors.New("email is required")
		}
	default:
		return errors.New("unsupported channel type")
	}
	return nil
}
