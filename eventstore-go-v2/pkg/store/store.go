package store

import (
	"github.com/pix303/eventstore-go-v2/internal/repository"
	"github.com/pix303/eventstore-go-v2/pkg/domain"
)

type EventStoreRepository interface {
	Append(event domain.StoreEvent) (domain.StoreEvent, error)
	RetriveByID(id string) (domain.StoreEvent, bool, error)
	RetriveByAggregateID(id string) ([]domain.StoreEvent, bool, error)
	RetriveByAggregateName(name string) ([]domain.StoreEvent, bool, error)
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

func (store *EventStore) Add(event domain.StoreEvent) (domain.StoreEvent, error) {
	result, err := store.repository.Append(event)
	return result, err
}

func (store *EventStore) GetByName(aggregateName string) ([]domain.StoreEvent, error) {
	result, ok, err := store.repository.RetriveByAggregateName(aggregateName)
	if ok {
		return result, nil
	}
	return []domain.StoreEvent{}, err
}

func (store *EventStore) GetByID(aggregateID string) ([]domain.StoreEvent, error) {
	result, ok, err := store.repository.RetriveByAggregateID(aggregateID)
	if ok {
		return result, nil
	}
	return []domain.StoreEvent{}, err
}

func (store *EventStore) GetByEventID(ID string) (domain.StoreEvent, bool, error) {
	result, ok, err := store.repository.RetriveByID(ID)
	if ok {
		return result, ok, nil
	}
	return domain.StoreEvent{}, ok, err
}
