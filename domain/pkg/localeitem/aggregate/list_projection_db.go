package aggregate

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/postgres-util-go/pkg/postgres"
)

type LocaleItemAggregateListState struct {
	repository *sqlx.DB
	publisher  *nats.Conn
}

var LocaleItemAggregateListAddress = actor.NewAddress("local", "list-aggregate-persister")

func NewLocaleItemAggregateListState() (*LocaleItemAggregateListState, error) {
	db, err := postgres.NewPostgresqlRepository()
	if err != nil {
		return nil, err
	}

	natsToken := os.Getenv("NATS_SECRET")
	nc, err := nats.Connect(nats.DefaultURL, nats.Token(natsToken))
	if err != nil {
		return nil, err
	}

	return &LocaleItemAggregateListState{
		repository: db,
		publisher:  nc,
	}, nil
}

type AddLocaleItemAggregateListBody struct {
	Aggregate LocaleItemAggregate
}

type GetContextBody struct {
	Id string
}

type GetContextBodyResult struct {
	Items []LocaleItemList
}

func (state *LocaleItemAggregateListState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	case AddLocaleItemAggregateListBody:
		state.addHandler(payload.Aggregate)
	case GetContextBody:
		result, err := state.getList(payload.Id)
		if err != nil {
			slog.Error("error on persist list", slog.String("err", err.Error()))
		}
		if msg.WithReturn {
			returnMsg := actor.NewReturnMessage(GetContextBodyResult{Items: result}, msg)
			msg.ReturnChan <- actor.NewWrappedMessage(&returnMsg, err)
		}
	}
}

func (state *LocaleItemAggregateListState) addHandler(aggregate LocaleItemAggregate) {
	err := state.persistList(aggregate)
	if err != nil {
		slog.Error("error on persist list", slog.String("err", err.Error()))
		return
	}

	err = state.publisher.Publish("locale.list.context.updated", []byte(aggregate.Context))
	if err != nil {
		slog.Error("error on publish list updated", slog.String("err", err.Error()))
	}
}

var listitemInsertOrUpdate string = `INSERT INTO locale.localeitems_list (aggregate_id, lang, content, context, updated_at, updated_by, is_lang_reference)
VALUES (:aggregate_id, :lang, :content, :context, :updated_at, :updated_by, :is_lang_reference)
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
	err := state.repository.Close()
	if err != nil {
		slog.Error("fail to close repository", slog.String("error", err.Error()))
	}
	state.repository = nil

	state.publisher.Close()
	state.publisher = nil
}
