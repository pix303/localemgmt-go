package aggregate

import (
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/cinecity/pkg/batch"
	"log/slog"

	"github.com/pix303/eventstore-go-v2/pkg/store"
)

var (
	ErrToRetriveAggregateEvents = "error on retriving aggregate events"
	ErrToPersistAggregate       = "error on persisting aggregate"
)

// LocaleItemAggregateState is the actor state for the aggregate persistence
type LocaleItemAggregateState struct {
	store     *store.EventStore
	batcher   *batch.Batcher
	aggregate *LocaleItemAggregate
}

var LocaleItemAggregateAddress = actor.NewAddress("local", "aggregate")

func NewLocaleItemAggregateActor() (*actor.Actor, error) {

	// create event store reference
	es, err := store.NewEventStore([]store.EventStoreConfigurator{store.WithPostgresqlRepository})

	if err != nil {
		return nil, err
	}

	// create actor state
	aggregate := NewLocaleItemAggregate()
	s := LocaleItemAggregateState{
		store:     &es,
		aggregate: &aggregate,
	}

	a, err := actor.NewActor(
		actor.NewAddress("local", "localeitem-aggregate"),
		&s,
	)

	if err != nil {
		return nil, err
	}

	// subscribe event store notifies
	addSubMsg := actor.NewAddSubcriptionMessage(a.GetAddress(), store.EventStoreAddress)
	err = actor.SendMessage(addSubMsg)

	if err != nil {
		return nil, err
	}

	b := batch.NewBatcher(5000, 5, s.updateAggregateState)
	s.batcher = b

	detailPersistState, err := NewLocaleItemAggregateDetailState()
	if err != nil {
		return nil, err
	}
	detailPersisterActor, err := actor.NewActor(
		LocaleItemAggregateDetailAddress,
		detailPersistState,
	)
	if err != nil {
		return nil, err
	}
	err = actor.RegisterActor(&detailPersisterActor)
	if err != nil {
		return nil, err
	}

	listPersistState, err := NewLocaleItemAggregateListState()
	if err != nil {
		return nil, err
	}
	listPersistActor, err := actor.NewActor(
		LocaleItemAggregateListAddress,
		listPersistState,
	)
	if err != nil {
		return nil, err
	}
	err = actor.RegisterActor(&listPersistActor)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (state *LocaleItemAggregateState) Process(msg actor.Message) {
	switch msg.Body.(type) {
	case store.StoreEventAddedBody:
		state.batcher.Add(msg)
	}
}

func (state *LocaleItemAggregateState) updateAggregateState(msg actor.Message) {
	body, ok := msg.Body.(store.StoreEventAddedBody)
	if !ok {
		slog.Warn("can not manage different message with LocaleItemAggregateDetailEventBody body: ignored")
		return
	}

	// TODO: use actor??
	evts, _, err := state.store.Repository.RetriveByAggregateID(body.AggregateID)
	if err != nil {
		slog.Warn(ErrToRetriveAggregateEvents, slog.String("error", err.Error()))
		return
	}

	newAgg := NewLocaleItemAggregate()
	state.aggregate = &newAgg
	state.aggregate.Reduce(evts)

	detailMsg := actor.NewMessage(
		LocaleItemAggregateDetailAddress,
		LocaleItemAggregateAddress,
		AddLocaleItemAggregateDetailBody{*state.aggregate},
		false,
	)

	listMsg := actor.NewMessage(
		LocaleItemAggregateListAddress,
		LocaleItemAggregateAddress,
		AddLocaleItemAggregateListBody{*state.aggregate},
		false,
	)

	err = actor.SendMessage(detailMsg)
	if err != nil {
		slog.Error(ErrToPersistAggregate, slog.String("error", err.Error()))
	}

	err = actor.SendMessage(listMsg)
	if err != nil {
		slog.Error(ErrToPersistAggregate, slog.String("error", err.Error()))
	}
}

func (state *LocaleItemAggregateState) GetState() any {
	return state.aggregate
}

func (state *LocaleItemAggregateState) Shutdown() {
	state.store = nil
	state.batcher = nil
	state.aggregate = nil
}
