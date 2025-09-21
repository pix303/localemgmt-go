package aggregate

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/pix303/eventstore-go-v2/pkg/events"
	"github.com/pix303/eventstore-go-v2/pkg/utils"
	domain "github.com/pix303/localemgmt-go/domain/pkg/localeitem/events"
)

const DEFAULT_CONTEXT = "default"

type TranslationItem struct {
	Lang      string
	Content   string
	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}

func NewTranslationItem(lang, content, userId string) TranslationItem {
	return TranslationItem{
		Lang:      lang,
		Content:   content,
		CreatedBy: userId,
		UpdatedBy: userId,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func (item *TranslationItem) UpdateTranslationItem(lang, content, userId string) {
	item.Lang = lang
	item.Content = content
	item.UpdatedBy = userId
	item.UpdatedAt = time.Now().UTC()
}

type LocaleItemAggregate struct {
	AggregateID   string
	Context       string
	ReferenceLang string
	Translations  []TranslationItem
}

const EMPTY_ID = "no-id"
const EMPTY_CONTEXT = "no-context"

func NewLocaleItemAggregate() LocaleItemAggregate {
	return LocaleItemAggregate{
		EMPTY_ID,
		EMPTY_CONTEXT,
		"",
		make([]TranslationItem, 0),
	}
}

func (a *LocaleItemAggregate) GetTranslationItemByLang(lang string) (*TranslationItem, error) {
	for _, t := range a.Translations {
		if t.Lang == lang {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("lang %s do not exist", lang)
}

func (a *LocaleItemAggregate) Reduce(evts []events.StoreEvent) {
	for _, evt := range evts {
		a.Apply(evt)
	}
}

func (a *LocaleItemAggregate) Apply(evt events.StoreEvent) {
	switch evt.EventType {
	case domain.CreateLocaleItemStoreEventType:
		a.init(evt)
	case domain.UpdateTranslationStoreEventType:
		a.update(evt)
	}
}

func (item *LocaleItemAggregate) init(evt events.StoreEvent) {
	createPayloadEvent, err := utils.DecodePayload[domain.CreateLocaleItemPayload](evt.PayloadData)
	if err != nil {
		slog.Error("error on decode payload", slog.String("payloadDataType", evt.PayloadDataType))
	}
	item.AggregateID = evt.AggregateID
	item.Context = createPayloadEvent.Context
	item.ReferenceLang = createPayloadEvent.Lang
	item.Translations = append(item.Translations, NewTranslationItem(
		createPayloadEvent.Lang,
		createPayloadEvent.Content,
		createPayloadEvent.CreatedBy,
	),
	)
}

func (item *LocaleItemAggregate) update(evt events.StoreEvent) {
	updatePayloadEvent, err := utils.DecodePayload[domain.UpdateTranslationLocaleItemPayload](evt.PayloadData)

	if err != nil {
		slog.Error("error on decode payload", slog.String("payloadDataType", evt.PayloadDataType))
	}

	langFounded := false
	for i := 0; i < len(item.Translations); i++ {
		t := &item.Translations[i]
		if t.Lang == updatePayloadEvent.Lang {
			t.Content = updatePayloadEvent.Content
			t.UpdatedAt = time.Now().UTC()
			t.UpdatedBy = evt.CreatedBy
			langFounded = true
			break
		}
	}

	if !langFounded {
		nt := NewTranslationItem(updatePayloadEvent.Lang, updatePayloadEvent.Content, "todo")
		slog.Info("new translation item", slog.Any("translation", nt))
		item.Translations = append(item.Translations, nt)
	}
}
