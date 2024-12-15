package domain

import (
	"fmt"
	"time"

	"github.com/beevik/guid"
)

type StoreEvent struct {
	EventType     string
	ID            string
	AggregateID   string
	AggregateName string
	CreatedAt     time.Time
	CreatedBy     string
}

func NewStoreEvent(eventType string, aggregateName string) StoreEvent {
	se := StoreEvent{
		EventType:     eventType,
		ID:            guid.New().String(),
		AggregateID:   guid.New().String(),
		AggregateName: aggregateName,
		CreatedAt:     time.Now(),
		CreatedBy:     "no-user",
	}

	return se
}

func NewDefaultStoreEvent() StoreEvent {
	se := StoreEvent{
		EventType:     "no-type",
		ID:            "",
		AggregateID:   "",
		AggregateName: "no-name",
		CreatedAt:     time.Now(),
		CreatedBy:     "no-user",
	}

	return se
}

func (event StoreEvent) ToString() string {
	return fmt.Sprintf("type %s, for aggregate %s (%s)", event.EventType, event.AggregateName, event.AggregateID)
}
