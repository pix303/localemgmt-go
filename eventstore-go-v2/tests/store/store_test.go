package store_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/pix303/eventstore-go-v2/pkg/domain"
	"github.com/pix303/eventstore-go-v2/pkg/store"
)

func exitWithError(err error) {
	if err != nil {
		fmt.Println(fmt.Errorf("exit for %v", err))
		os.Exit(1)
	}
}

func TestStore_ok(t *testing.T) {
	store, err := store.NewEventStore(store.WithInMemoryRepository)
	exitWithError(err)

	evt := domain.NewStoreEvent("some-event", "something")

	storedEvent, err := store.Add(evt)
	exitWithError(err)

	if storedEvent.EventType != "some-event" {
		exitWithError(errors.New(fmt.Sprintf("expect some-event, got %s", storedEvent.EventType)))
	}

	_, err = store.Add(evt)
	_, err = store.Add(evt)
	exitWithError(err)

	result, err := store.GetByName("something")
	exitWithError(err)

	if len(result) != 3 {
		exitWithError(errors.New(fmt.Sprintf("expect 3 length, got %d", len(result))))
	}

	result, err = store.GetByID(evt.AggregateID)
	exitWithError(err)
	if len(result) != 3 {
		exitWithError(errors.New(fmt.Sprintf("expect 3 length, got %d", len(result))))
	}

	resultByID, ok, err := store.GetByEventID(evt.ID)
	exitWithError(err)

	if ok != true {
		exitWithError(errors.New(fmt.Sprintf("expect to be found , got %v", ok)))
	}

	if resultByID.AggregateName != "something" {
		t.Errorf("event aggregate name must be something istead of %s", resultByID.AggregateName)
		exitWithError(errors.New(fmt.Sprintf("expect to be something, got %v", resultByID.AggregateName)))
	}
}
