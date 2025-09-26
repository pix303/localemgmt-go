package aggregate

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/postgres-util-go/pkg/postgres"
)

type LocaleItemAggregateListState struct {
	repository *sqlx.DB
}

var LocaleItemAggregateListAddress = actor.NewAddress("local", "list-aggregate-persister")

func NewLocaleItemAggregateListState() (*LocaleItemAggregateListState, error) {
	db, err := postgres.NewPostgresqlRepository()
	if err != nil {
		return nil, err
	}

	return &LocaleItemAggregateListState{
		repository: db,
	}, nil
}

type AddLocaleItemAggregateListBody struct {
	Aggregate LocaleItemAggregate
}

func (state *LocaleItemAggregateListState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	case AddLocaleItemAggregateListBody:
		err := state.persistList(payload.Aggregate)
		if err != nil {
			slog.Error("error on persist list", slog.String("err", err.Error()))
		}
	case GetContextBody:
		result, err := state.getList(payload.Id)
		if err != nil {
			slog.Error("error on persist list", slog.String("err", err.Error()))
		}
		returnMsg := actor.NewReturnMessage(GetContextBodyResult{Items: result}, msg)
		if msg.WithReturn {
			msg.ReturnChan <- actor.NewWrappedMessage(&returnMsg, err)
		}
	}
}

var listitemInsertOrUpdate string = `INSERT INTO locale.localeitems_list (aggregate_id, lang, content, context, updated_at, updated_by, is_lang_reference)
VALUES (:id, :lang, :content, :context, :updated_at, :updated_by, :is_lang_heference)
ON CONFLICT (aggregate_id, lang )
DO UPDATE SET
    content = :content,
    updated_at = :updated_at,
    updated_by = :updated_by,
    is_lang_reference = :is_lang_reference;
`

func (state *LocaleItemAggregateListState) persistList(aggregate LocaleItemAggregate) error {

	slog.Debug("start insert or update aggregate translations in list projection")
	tx, err := state.repository.Beginx()
	if err != nil {
		slog.Error("fail to begin transaction on db", slog.String("error", err.Error()))
		return fmt.Errorf("failed to begin transaction %w", err)
	}

	defer func() {
		if err != nil {
			if rberr := tx.Rollback(); rberr != nil {
				slog.Error("transation fail so apply rollback", slog.String("error", err.Error()))
			}
		}
	}()

	for _, tItem := range aggregate.Translations {
		user := tItem.UpdatedBy
		if user == "" {
			user = tItem.CreatedBy
			if user == "" {
				user = "TODO"
			}
		}

		params := NewLocaleItemList(
			aggregate.AggregateID,
			tItem.Lang,
			tItem.Content,
			aggregate.Context,
			tItem.UpdatedAt,
			user,
			aggregate.ReferenceLang == tItem.Lang,
		)
		_, err = tx.NamedExec(listitemInsertOrUpdate, params)

		if err != nil {
			slog.Error("fail insert/update statement",
				slog.String("lang", tItem.Lang),
				slog.String("content", tItem.Content),
				slog.String("id", aggregate.AggregateID),
				slog.String("err", err.Error()),
			)
		}
	}

	slog.Debug("finish insert or update aggregate translations in list projection")
	return tx.Commit()
}

func (state *LocaleItemAggregateListState) getList(context string) ([]LocaleItemList, error) {
	result := make([]LocaleItemList, 0)
	err := state.repository.Select(&result, "SELECT * FROM locale.localeitems_list WHERE context = $1", context)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (state *LocaleItemAggregateListState) GetState() any {
	return nil
}

func (state *LocaleItemAggregateListState) Shutdown() {
	state.repository = nil
}
