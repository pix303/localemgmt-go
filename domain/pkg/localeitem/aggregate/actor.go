package aggregate

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/cinecity/pkg/batch"

	"github.com/pix303/eventstore-go-v2/pkg/store"
	"github.com/pix303/postgres-util-go/pkg/postgres"
)

var (
	ErrToRetriveAggregateEvents = "error on retriving aggregate events"
	ErrToPersistAggregate       = "error on persisting aggregate"
)

type LocaleItemAggregatePersister interface {
	Persist(aggregate LocaleItemAggregate) error
}

type LocaleItemAggregateLocalPersisterOnFile struct{}

const LocaleItemAggregateLocalFolder = "./localeitems-bucket"

func (persister *LocaleItemAggregateLocalPersisterOnFile) Persist(aggregate LocaleItemAggregate) error {
	slog.Info("start persisting aggregate", slog.String("aggregateId", aggregate.AggregateID))
	aggregateDir, err := os.Open(LocaleItemAggregateLocalFolder)
	if err != nil {
		err = os.Mkdir(LocaleItemAggregateLocalFolder, 0755)
		if err != nil {
			return fmt.Errorf("error on create local folder for aggregates: %s", err.Error())
		}
		aggregateDir, _ = os.Open(LocaleItemAggregateLocalFolder)
	}

	// resjson, err := json.MarshalIndent(aggregate, "", "  ")
	resjson, err := json.Marshal(aggregate)
	if err != nil {
		slog.Warn("fail to marshal aggregate", slog.String("error", err.Error()))
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s.json", aggregateDir.Name(), aggregate.AggregateID), resjson, 0755)
	if err != nil {
		slog.Warn("fail to write aggregate file", slog.String("error", err.Error()))
		return err
	}
	return nil
}

type LocaleItemAggregatePersisterOnDB struct {
	repository *sqlx.DB
}

func NewLocaleItemAggregatePersisterOnDB() (*LocaleItemAggregatePersisterOnDB, error) {
	db, err := postgres.NewPostgresqlRepository()
	if err != nil {
		return nil, err
	}

	return &LocaleItemAggregatePersisterOnDB{
		repository: db,
	}, nil
}

var insertOrUpdate string = `INSERT INTO locale.localeitems (aggregateId, updatedAt, data)
VALUES (:id, :upDate, :data)
ON CONFLICT (aggregateId) 
DO UPDATE SET 
    updatedAt = :upDate,
    data = :data;
`

func (persiter *LocaleItemAggregatePersisterOnDB) Persist(aggregate LocaleItemAggregate) error {
	datajson, err := json.Marshal(aggregate)
	if err != nil {
		slog.Warn("fail to marshal aggregate", slog.String("error", err.Error()))
		return err
	}

	res, err := persiter.repository.NamedExec(insertOrUpdate, map[string]any{
		"id":     aggregate.AggregateID,
		"data":   string(datajson),
		"upDate": time.Now(),
	})

	if err != nil {
		slog.Warn("fail to persist aggregate", slog.String("error", err.Error()))
		return err
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		slog.Warn("fail to persist aggregate", slog.String("error", err.Error()))
		return err
	}

	if numRows != 1 {
		return fmt.Errorf("no row updated or interted for %s", aggregate.AggregateID)
	}

	return nil
}

type LocaleItemAggregateState struct {
	store     *store.EventStore
	batcher   *batch.Batcher
	persister LocaleItemAggregatePersister
}

func NewLocaleItemAggregateActor() (*actor.Actor, error) {
	// TODO: use store actor??
	// create event store reference
	es, err := store.NewEventStore([]store.EventStoreConfigurator{store.WithPostgresqlRepository})

	if err != nil {
		return nil, err
	}

	// create actor state
	s := LocaleItemAggregateState{
		store: &es,
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
	err = a.Send(addSubMsg)

	if err != nil {
		return nil, err
	}

	b := batch.NewBatcher(5000, 5, s.updateAggregateState)
	s.batcher = b

	persister, err := NewLocaleItemAggregatePersisterOnDB()
	if err != nil {
		return nil, err
	}

	s.persister = persister
	return &a, nil
}

// type LocaleItemAggregateDetailEventBody struct {
// 	AggregateID string
// 	EventType   string
// 	Payload     string
// }

func (state *LocaleItemAggregateState) Process(inbox <-chan actor.Message) {
	for {
		msg := <-inbox
		switch msg.Body.(type) {
		case store.StoreEventAddedBody:
			state.batcher.Add(msg)
		}
	}
}

func (state *LocaleItemAggregateState) updateAggregateState(msg actor.Message) {
	body, ok := msg.Body.(store.StoreEventAddedBody)
	if !ok {
		slog.Warn("cant manage different message with LocaleItemAggregateDetailEventBody body: ignored")
		return
	}
	// TODO: use actor??
	evts, _, err := state.store.Repository.RetriveByAggregateID(body.AggregateID)
	if err != nil {
		slog.Warn(ErrToRetriveAggregateEvents, slog.String("error", err.Error()))
		return
	}

	aggregate := NewLocaleItemAggregate()
	aggregate.Reduce(evts)

	err = state.persister.Persist(aggregate)
	if err != nil {
		slog.Warn(ErrToPersistAggregate, slog.String("error", err.Error()))
	}
}

func (state *LocaleItemAggregateState) Shutdown() {
	state.store = nil
	state.batcher = nil
	state.persister = nil
}
