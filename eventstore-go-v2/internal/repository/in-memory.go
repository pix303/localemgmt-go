package repository

import (
	"errors"

	"github.com/pix303/eventstore-go-v2/pkg/domain"
)

type InMemoryRepository struct {
	events []domain.StoreEvent
}

func (repo *InMemoryRepository) Append(event domain.StoreEvent) (domain.StoreEvent, error) {
	repo.events = append(repo.events, event)
	return event, nil
}

func (repo *InMemoryRepository) RetriveByID(id string) (domain.StoreEvent, bool, error) {
	for _, evt := range repo.events {
		if evt.ID == id {
			return evt, true, nil
		}
	}
	return domain.StoreEvent{}, false, errors.New("not found")
}

func (repo *InMemoryRepository) RetriveByAggregateID(id string) ([]domain.StoreEvent, bool, error) {
	result := []domain.StoreEvent{}
	for _, evt := range repo.events {
		if evt.AggregateID == id {
			result = append(result, evt)
		}
	}

	if len(result) > 0 {
		return result, true, nil
	}
	return []domain.StoreEvent{}, false, errors.New("not found")
}

func (repo *InMemoryRepository) RetriveByAggregateName(name string) ([]domain.StoreEvent, bool, error) {
	result := []domain.StoreEvent{}
	for _, evt := range repo.events {
		if evt.AggregateName == name {
			result = append(result, evt)
		}
	}

	if len(result) > 0 {
		return result, true, nil
	}
	return []domain.StoreEvent{}, false, errors.New("not found")
}
