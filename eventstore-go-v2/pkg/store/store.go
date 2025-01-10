package store

import (
	"github.com/pix303/eventstore-go-v2/internal/repository"
	"github.com/pix303/eventstore-go-v2/pkg/events"
)

type EventStoreRepository interface {
	Append(event events.AggregateEvent) (bool, error)
	RetriveByID(id string) (*events.AggregateEvent, bool, error)
	RetriveByAggregateID(id string) ([]events.AggregateEvent, bool, error)
	RetriveByAggregateName(name string) ([]events.AggregateEvent, bool, error)
}

type EventStore struct {
	repository EventStoreRepository
}

type EventStoreConfig func(store *EventStore) error

func NewEventStore(configure EventStoreConfig) (EventStore, error) {
	store := EventStore{}
	err := configure(&store)
	return store, err
}

func WithInMemoryRepository(store *EventStore) error {
	store.repository = &repository.InMemoryRepository{}
	return nil
}

func (store *EventStore) Add(event events.AggregateEvent) (bool, error) {
	result, err := store.repository.Append(event)
	return result, err
}

func (store *EventStore) GetByName(aggregateName string) ([]events.AggregateEvent, error) {
	result, ok, err := store.repository.RetriveByAggregateName(aggregateName)
	if ok {
		return result, nil
	}
	return []events.AggregateEvent{}, err
}

func (store *EventStore) GetByID(aggregateID string) ([]events.AggregateEvent, error) {
	result, ok, err := store.repository.RetriveByAggregateID(aggregateID)
	if ok {
		return result, nil
	}
	return []events.AggregateEvent{}, err
}

func (store *EventStore) GetByEventID(ID string) (*events.AggregateEvent, bool, error) {
	result, ok, err := store.repository.RetriveByID(ID)
	if ok {
		return result, ok, nil
	}
	return nil, ok, err
}
