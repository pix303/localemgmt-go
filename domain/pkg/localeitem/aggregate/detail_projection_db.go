package aggregate

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/postgres-util-go/pkg/postgres"
)

type LocaleItemAggregateDetailState struct {
	repository *sqlx.DB
}

var LocaleItemAggregateDetailAddress = actor.NewAddress("local", "detail-aggregate-persister")

func NewLocaleItemAggregateDetailState() (*LocaleItemAggregateDetailState, error) {
	db, err := postgres.NewPostgresqlRepository()
	if err != nil {
		return nil, err
	}

	return &LocaleItemAggregateDetailState{
		repository: db,
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

func (state *LocaleItemAggregateDetailState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	case AddLocaleItemAggregateDetailBody:
		err := state.persistDetail(payload.Aggregate)
		if err != nil {
			slog.Error("error on persist detail", slog.String("err", err.Error()))
		}
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
