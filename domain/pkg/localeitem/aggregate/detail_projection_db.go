package aggregate

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/postgres-util-go/pkg/postgres"
)

type LocaleItemAggregateDetailState struct {
	repository *sqlx.DB
	publisher  *nats.Conn
}

var LocaleItemAggregateDetailAddress = actor.NewAddress("local", "detail-aggregate-persister")

func NewLocaleItemAggregateDetailState() (*LocaleItemAggregateDetailState, error) {
	db, err := postgres.NewPostgresqlRepository()
	if err != nil {
		return nil, err
	}

	natsToken := os.Getenv("NATS_SECRET")
	nc, err := nats.Connect(nats.DefaultURL, nats.Token(natsToken))
	if err != nil {
		return nil, err
	}

	return &LocaleItemAggregateDetailState{
		repository: db,
		publisher:  nc,
	}, nil
}

var detailInsertOrUpdate string = `INSERT INTO locale.localeitem_detail (aggregateId, updatedAt, data)
VALUES (:id, :upDate, :data)
ON CONFLICT (aggregateId) 
DO UPDATE SET 
    updatedAt = :upDate,
    data = :data;
`

type AddLocaleItemAggregateDetailBody struct {
	Aggregate LocaleItemAggregate
}

type GetLocaleItemAggregateDetailBody struct {
	Id string
}

type GetLocaleItemAggregateDetailBodyResult struct {
	Aggregate LocaleItemAggregate
}

type GetContextBody struct {
	Id string
}

type GetContextBodyResult struct {
	Items []LocaleItemList
}

func (state *LocaleItemAggregateDetailState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	case AddLocaleItemAggregateDetailBody:
		state.addDetail(payload.Aggregate)
	case GetLocaleItemAggregateDetailBody:
		result, err := state.getDetail(payload.Id)
		if err != nil {
			slog.Error("error on get detail", slog.String("err", err.Error()))
		}

		resultMsg := actor.NewMessage(
			msg.From,
			msg.To,
			result,
			false,
		)

		if msg.WithReturn {
			msg.ReturnChan <- actor.NewWrappedMessage(&resultMsg, err)
		}
	}
}

func (state *LocaleItemAggregateDetailState) addDetail(aggregate LocaleItemAggregate) {
	slog.Info("----on persist detail")
	err := state.persistDetail(aggregate)
	if err != nil {
		slog.Error("error on persist detail", slog.String("err", err.Error()))
		return
	}

	err = state.publisher.Publish("locale.detail.updated", []byte(aggregate.AggregateID))
	if err != nil {
		slog.Error("error on publish detail updated", slog.String("err", err.Error()))
	}
}

func (state *LocaleItemAggregateDetailState) persistDetail(aggregate LocaleItemAggregate) error {
	datajson, err := json.Marshal(aggregate)
	if err != nil {
		slog.Warn("fail to marshal aggregate",
			slog.String("error", err.Error()),
			slog.String("aggregateId", aggregate.AggregateID),
		)
		return err
	}

	res, err := state.repository.NamedExec(detailInsertOrUpdate, map[string]any{
		"id":     aggregate.AggregateID,
		"data":   string(datajson),
		"upDate": time.Now().UTC(),
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

const selectDetailByID = `SELECT data FROM locale.localeitem_detail WHERE aggregateid = $1`

func (state *LocaleItemAggregateDetailState) getDetail(id string) (LocaleItemAggregate, error) {
	result := LocaleItemAggregate{}
	row, err := state.repository.Query(selectDetailByID, id)
	if err != nil {
		return result, err
	}
	defer row.Close()

	if row.Err() != nil {
		return result, row.Err()
	}

	for row.Next() {
		var data string
		err = row.Scan(&data)
		if err != nil {
			return result, err
		}
		err = json.Unmarshal([]byte(data), &result)
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

func (state *LocaleItemAggregateDetailState) GetState() any {
	return nil
}

func (state *LocaleItemAggregateDetailState) Shutdown() {
	state.repository = nil
}
