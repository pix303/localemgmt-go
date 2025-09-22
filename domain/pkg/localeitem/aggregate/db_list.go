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

type LocaleItemAggregateListBody struct {
	Aggregate LocaleItemAggregate
}

func (state *LocaleItemAggregateListState) Process(inbox <-chan actor.Message) {
	for {
		msg := <-inbox
		switch payload := msg.Body.(type) {
		case LocaleItemAggregateListBody:
			err := state.persist(payload.Aggregate)
			if err != nil {
				slog.Error("error on persist list", slog.String("err", err.Error()))
			}
		}
	}
}

var listitemInsertOrUpdate string = `INSERT INTO locale.localeitems_list (aggregate_id, lang, content, context, updated_at, updated_by, is_lang_reference)
VALUES (:id, :lang, :content, :context, :updatedAt, :updatedBy, :isLangReference)
ON CONFLICT (aggregate_id, lang )
DO UPDATE SET
    content = :content,
    updated_at = :updatedAt,
    updated_by = :updatedBy,
    is_lang_reference = :isLangReference;
`

func (state *LocaleItemAggregateListState) persist(aggregate LocaleItemAggregate) error {

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

		_, err = tx.NamedExec(listitemInsertOrUpdate,
			map[string]any{
				"id":              aggregate.AggregateID,
				"content":         tItem.Content,
				"context":         aggregate.Context,
				"lang":            tItem.Lang,
				"updatedAt":       tItem.UpdatedAt,
				"updatedBy":       user,
				"isLangReference": aggregate.ReferenceLang == tItem.Lang,
			},
		)

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

func (state *LocaleItemAggregateListState) Shutdown() {
	state.repository = nil
}
